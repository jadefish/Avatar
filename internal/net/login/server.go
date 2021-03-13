package login

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/jadefish/avatar/pkg/commands"
	"github.com/jadefish/avatar/pkg/crypto"
	"github.com/jadefish/avatar/pkg/encoding/binary"

	"github.com/jadefish/avatar"
	net2 "github.com/jadefish/avatar/pkg/net"
)

// TODO: remove these when shards are loaded from a shard service.
var tz *time.Location
var tmpIP net.IP

func init() {
	// TODO: remove `init` when shards are loaded from a shard service.
	var err error
	if tz, err = time.LoadLocation("America/Detroit"); err != nil {
		panic(err)
	}

	if tmpIP = net.ParseIP("192.168.0.20"); tmpIP == nil {
		panic("unable to parse tmpIP")
	} else {
		tmpIP = tmpIP.To4()
	}
}

// Server is a capable of accepting and processing client login requests over a
// network.
type Server struct {
	Accounts *avatar.AccountService
	log      *log.Logger

	addr     string
	listener net.Listener

	cm      sync.RWMutex
	clients map[*net2.Client]struct{}

	errs chan error

	stop chan bool
}

// NewServer creates a new Login server.
func NewServer(
	accountService *avatar.AccountService,
	addr string,
	log *log.Logger,
) *Server {
	return &Server{
		Accounts: accountService,
		log:      log,
		addr:     addr,
		stop:     make(chan bool),
		clients:  make(map[*net2.Client]struct{}, 100),
		errs:     make(chan error),
	}
}

func loginDeniedReason(err error) binary.LoginDeniedReason {
	if errors.Is(err, avatar.ErrInvalidCredentials) {
		return binary.LoginDeniedReasonUnableToAuthenticate
	} else if errors.Is(err, avatar.ErrAccountNotFound) {
		return binary.LoginDeniedReasonUnableToAuthenticate
	} else if errors.Is(err, avatar.ErrAccountInUse) {
		return binary.LoginDeniedReasonAccountInUse
	} else if errors.Is(err, avatar.ErrAccountBlocked) {
		return binary.LoginDeniedReasonAccountBlocked
	} else {
		return binary.LoginDeniedReasonCommunicationProblem
	}

	// TODO: server at capacity, game server down, timeouts?
}

func (s *Server) denyLogin(client *net2.Client, err error) error {
	if errors.Is(err, io.EOF) {
		// if server received EOF, just terminate the connection.
		s.disconnectClient(client)

		return nil
	}

	s.errs <- err
	cmd := &binary.LoginDenied{Reason: loginDeniedReason(err)}

	if err := client.SendCommand(cmd); err != nil {
		return fmt.Errorf("deny login: %w", err)
	}

	s.disconnectClient(client)

	return nil
}

func (s *Server) addClient(client *net2.Client) error {
	s.cm.Lock()
	defer s.cm.Unlock()

	for c := range s.clients {
		if c == client || c.Identifier() == client.Identifier() {
			return avatar.ErrAccountInUse
		}
	}

	s.clients[client] = struct{}{}

	return nil
}

func (s *Server) disconnectClient(client *net2.Client) {
	s.cm.Lock()
	defer s.cm.Unlock()

	if err := client.Disconnect(); err != nil {
		s.errs <- fmt.Errorf("disconnect client: %w", err)
	}

	s.log.Println("disconnected client", client.Addr())
	delete(s.clients, client)
}

func (s *Server) authenticate(client *net2.Client) error {
	// 1. ReceiveCommand Login Seed:
	loginSeed := &binary.LoginSeed{}
	if err := client.ReceiveCommand(loginSeed); err != nil {
		return fmt.Errorf("authenticate client: %w", err)
	}

	cs, err := crypto.NewLoginCryptoService(loginSeed.Seed, &loginSeed.Version)

	if err != nil {
		return fmt.Errorf("authenticate client: %w", err)
	}

	client.SetCryptoService(cs)

	// 2. ReceiveCommand Login Request:
	loginRequest := &binary.LoginRequest{}
	if err := client.ReceiveCommand(loginRequest); err != nil {
		return fmt.Errorf("authenticate client: %w", err)
	}

	// 3. Authenticate provided credentials:
	_, err = commands.Authenticate{Accounts: s.Accounts}.
		Call(loginRequest.AccountName, loginRequest.Password)

	if err != nil {
		return fmt.Errorf("authenticate client: %w", err)
	}

	// TODO: client.Account = account ???
	client.SetState(avatar.ClientStateAuthenticated)

	if err := s.addClient(client); err != nil {
		return fmt.Errorf("authenticate client: %w", err)
	}

	// 4. SendCommand Game Server List:
	// TODO: populate list from some shard service
	gameServerList := &binary.GameServerList{SystemInfoFlag: binary.SystemInfoFlagUnknown}
	gameServerList.AddItem(&binary.GameServerListItem{
		Index:       0,
		Name:        "Foo",
		PercentFull: 10,
		TimeZone:    tz,
		IP:          tmpIP,
	})

	if err := client.SendCommand(gameServerList); err != nil {
		return fmt.Errorf("authenticate client: %w", err)
	}

	return nil
}

// relay a client to the game server indicated by the received "Connect to Game
// Server" packet.
func (s *Server) relay(client *net2.Client) error {
	// 0. recv 268 bytes: 0xD9 Client Info (E)
	//    client does not always send this packet.
	// 1. recv   3 bytes: 0xA0 Select Game Server (E)
	// 2. send  11 bytes: 0x8C Connect to Game Server
	buf := make([]byte, net2.MaxPacketSize)
	n, err := client.Read(buf)

	if err != nil {
		return fmt.Errorf("relay client: %w", err)
	}

	buf = buf[:n]

	if err := client.Decrypt(buf, buf); err != nil {
		return fmt.Errorf("relay client: %w", err)
	}

	switch buf[0] {
	case 0xD9:
		// client sent optional packet 0xD9 Client Info.
		// TODO: do something with Client Info?

		// prep for next packet:
		buf = make([]byte, 3)
		n, err = client.Read(buf)

		if err != nil {
			return fmt.Errorf("relay client: %w", err)
		}

		if err := client.Decrypt(buf, buf); err != nil {
			return fmt.Errorf("relay client: %w", err)
		}

		fallthrough
	case 0xA0:
		sgs := &binary.SelectServer{}
		if err := sgs.UnmarshalBinary(net2.ByteOrder, buf); err != nil {
			return fmt.Errorf("relay client: %w", err)
		}

		// TODO: match up sgs.Index with game server list index
		connectToGameServer := &binary.ConnectToGameServer{
			IP:     tmpIP,
			Port:   7780,
			NewKey: uint32(0),
		}

		if err := client.SendCommand(connectToGameServer); err != nil {
			return fmt.Errorf("relay client: %w", err)
		}
	default:
		return fmt.Errorf("relay client: unexpected packet 0x%X", buf[0])
	}

	return nil
}

// Start the server, allowing it to accept and process incoming client
// connections.
func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)

	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s.log.Println("listening on", s.Address())

	go func() {
		for {
			select {
			case err := <-s.errs:
				if err != nil {
					s.log.Println(err)
				}
			}
		}
	}()

	for {
		if conn, err := s.listener.Accept(); err != nil {
			select {
			case <-s.stop:
				return nil
			default:
				s.errs <- fmt.Errorf("accept: %w", err)
			}
		} else {
			go s.handleConn(conn)
		}
	}
}

func (s *Server) handleConn(conn net.Conn) {
	s.log.Println("new connection from", conn.RemoteAddr())
	client := net2.NewClient(conn)
	// _ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	for {
		switch client.GetState() {
		default:
			s.disconnectClient(client)
			return
		case avatar.ClientStateNew:
			// New clients need to be authenticated.
			if err := s.authenticate(client); err != nil {
				s.errs <- s.denyLogin(client, err)
				return
			}
		case avatar.ClientStateAuthenticated:
			// Authenticated clients need to be relayed to the game server they
			// have selected.
			if err := s.relay(client); err != nil {
				s.errs <- s.denyLogin(client, err)
				return
			}

			// Client has been relayed and has established a new connection to
			// its selected destination game server.
			s.disconnectClient(client)
			return
		}
	}
}

// Stop the server, disconnecting connected clients and preventing new
// connections from being accepted.
func (s *Server) Stop() error {
	s.log.Println("Stopping server...")

	for client := range s.clients {
		s.disconnectClient(client)
	}

	close(s.stop)
	close(s.errs)

	return s.listener.Close()
}

// Address retrieves the server's address.
func (s *Server) Address() string {
	return s.listener.Addr().String()
}
