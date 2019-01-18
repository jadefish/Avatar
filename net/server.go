package net

import (
	"log"
	"net"
	"os"

	"github.com/jadefish/avatar"
	"github.com/jadefish/avatar/command"
	"github.com/jadefish/avatar/packet"
	"github.com/pkg/errors"
)

// TODO: load me from separate configuration:
const (
	clientSliceSize = 256
	defaultHost     = "0.0.0.0"
	defaultPort     = "7775"
	// processCapacity = 5
)

// Server is a capable of accepting and processing clients over a network.
type Server struct {
	avatar.Server

	accounts  avatar.AccountService
	passwords avatar.PasswordService
	shards    avatar.ShardService

	clients  []Client
	addr     net.Addr
	listener net.Listener
}

// NewServer creates a new server.
func NewServer(
	as avatar.AccountService,
	ps avatar.PasswordService,
	ss avatar.ShardService,
) *Server {
	return &Server{
		clients:   make([]Client, 0, clientSliceSize),
		accounts:  as,
		passwords: ps,
		shards:    ss,
	}
}

func (s Server) AccountService() avatar.AccountService {
	return s.accounts
}

func (s Server) PasswordService() avatar.PasswordService {
	return s.passwords
}

func (s Server) ShardService() avatar.ShardService {
	return s.shards
}

// Start the server, allowing it to accept and process incoming client
// connections.
func (s *Server) Start() error {
	l, err := net.Listen("tcp", getAddr())

	if err != nil {
		return errors.Wrap(err, "listen")
	}

	// results := make(chan *result)

	s.listener = l
	s.addr = l.Addr()

	log.Println("Listening on", s.Address())

	for {
		conn, err := s.listener.Accept()

		if err != nil {
			return errors.Wrap(err, "accept")
		}

		go s.process(conn)

		// for r := range results {
		// 	if r.ok() {
		// 		s.addClient(*r.client)
		// 	} else {
		// 		log.Println(err)
		// 	}
		// }
	}
}

// Stop the server, disconnecting connected clients and preventing new
// connections from being accepted.
func (s *Server) Stop() error {
	log.Println("Stopping server...")

	// for _, c := range s.clients {
	// 	err := c.Disconnect(0x00) // TODO
	//
	// 	if err != nil {
	// 		log.Println(errors.Wrap(err, "server stop"))
	// 	}
	//
	// 	s.removeClient(&c)
	// }

	return s.listener.Close()
}

// Address retrieves the server's address.
func (s Server) Address() string {
	return s.listener.Addr().String()
}

func (s *Server) process(conn net.Conn) {

	client := NewClient(conn)

	for {
		// Read data into a packet:
		p, err := packet.New(client.conn, client.crypto)

		if err != nil {
			log.Println(err)
			conn.Close()
			return
		}

		cmd, err := command.Make(p.Descriptor(), p)

		if err != nil {
			log.Println(err)
			conn.Close()
			return
		}

		err = cmd.Execute(client, s)

		if err != nil {
			log.Println(err)
			conn.Close()
			return
		}

		log.Printf("client:\n+%v\n", client)
	}
}

func getAddr() string {
	if val, ok := os.LookupEnv("LOGIN_ADDR"); ok {
		if host, port, err := net.SplitHostPort(val); err == nil {
			return net.JoinHostPort(host, port)
		}
	}

	return net.JoinHostPort(defaultHost, defaultPort)
}
