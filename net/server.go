package net

import (
	"net"
	"os"

	"github.com/jadefish/avatar"
	"github.com/pkg/errors"
)

// TODO: load me from configuration:
const (
	clientSliceSize = 256
	defaultHost     = "0.0.0.0"
	defaultPort     = "7775"
)

// Server is capable of accepting and processing clients.
type Server struct {
	Accounts  avatar.AccountService
	Passwords avatar.PasswordService

	clients  []*Client
	addr     net.Addr
	listener net.Listener
}

// NewServer creates a new server.
func NewServer() *Server {
	return &Server{
		clients: make([]*Client, clientSliceSize),
	}
}

// Start the server.
func (s *Server) Start() error {
	l, err := net.Listen("tcp", getAddr())

	if err != nil {
		return err
	}

	s.listener = l
	s.addr = l.Addr()

	return nil
}

// Stop the server.
func (s *Server) Stop() error {
	return s.listener.Close()
}

// Address retrieves the server's address.
func (s *Server) Address() string {
	return s.listener.Addr().String()
}

// Accept a new connection, creating a client for the connection.
func (s *Server) Accept() (*Client, error) {
	conn, err := s.listener.Accept()

	if err != nil {
		return nil, errors.Wrap(err, "accept")
	}

	client, err := NewClient(conn, *s)

	if err != nil {
		return nil, errors.Wrap(err, "accept")
	}

	s.clients = append(s.clients, client)

	return client, nil
}

func (s *Server) FindClient(seed uint32) *Client {
	// for _, client := range s.clients {
	// 	if client.GetCrypto().GetSeed() == seed {
	// 		return client
	// 	}
	// }

	return nil
}

func (s *Server) GetClientsByState(state avatar.ClientState) []*Client {
	clients := []*Client{nil}

	// for _, client := range s.clients {
	// 	if client.GetState() == state {
	// 		clients = append(clients, client)
	// 	}
	// }

	return clients
}

func (s *Server) addClient(client *Client) error {
	s.clients = append(s.clients, client)

	return nil
}

func getAddr() string {
	if val, ok := os.LookupEnv("LOGIN_ADDR"); ok {
		if host, port, err := net.SplitHostPort(val); err == nil {
			return net.JoinHostPort(host, port)
		}
	}

	return net.JoinHostPort(defaultHost, defaultPort)
}
