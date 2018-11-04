package net

import (
	"log"
	"net"
	"os"

	"github.com/jadefish/avatar"
	"github.com/pkg/errors"
)

// TODO: load me from separate configuration:
const (
	clientSliceSize = 256
	defaultHost     = "0.0.0.0"
	defaultPort     = "7775"
)

// Server is a capable of accepting and processing clients over a network.
type Server struct {
	accounts  avatar.AccountService
	passwords avatar.PasswordService

	clients  []Client
	addr     net.Addr
	listener net.Listener
}

// NewServer creates a new server.
func NewServer(as avatar.AccountService, ps avatar.PasswordService) *Server {
	return &Server{
		clients:   make([]Client, clientSliceSize),
		accounts:  as,
		passwords: ps,
	}
}

// Start the server, allowing it to accept and process incoming client
// connections.
func (s *Server) Start() error {
	l, err := net.Listen("tcp", getAddr())

	if err != nil {
		return err
	}

	s.listener = l
	s.addr = l.Addr()

	log.Println("Listening on", s.Address())

	defer s.Stop()

	done := make(chan bool)
	errs := make(chan error)

	defer close(done)
	defer close(errs)

	for {
		client, err := s.accept()

		if err != nil {
			log.Println(errors.Wrap(err, "server accept"))
			continue
		}

		go s.processClient(client, done, errs)

		select {
		case <-done:
			s.addClient(*client)
		case e := <-errs:
			log.Println(errors.Wrap(e, "process client"))
			client.Disconnect(0x00) // TODO
		}
	}
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
func (s *Server) accept() (*Client, error) {
	conn, err := s.listener.Accept()

	if err != nil {
		return nil, errors.Wrap(err, "listener accept")
	}

	client, err := NewClient(conn, *s)

	if err != nil {
		return nil, errors.Wrap(err, "new client creation")
	}

	return client, nil
}

func (s Server) processClient(c *Client, done chan<- bool, errs chan<- error) {
	err := c.Connect()

	if err != nil {
		errs <- errors.Wrap(err, "process client")
		return
	}

	result, err := c.Authenticate()

	if err != nil {
		errs <- errors.Wrap(err, "process client")
		return
	}

	// Find account:
	account, err := s.accounts.GetAccountByName(result.AccountName)

	if err != nil {
		errs <- errors.Wrap(err, "process client")
		return
	}

	log.Println("found account:", account)

	// Authenticate:
	if !s.passwords.VerifyPassword(result.Password, []byte(account.Password)) {
		errs <- errors.Wrap(avatar.ErrInvalidCredentials, "process client")
		return
	}

	done <- true
}

func (s Server) addClient(client Client) {
	s.clients = append(s.clients, client)
}

func getAddr() string {
	if val, ok := os.LookupEnv("LOGIN_ADDR"); ok {
		if host, port, err := net.SplitHostPort(val); err == nil {
			return net.JoinHostPort(host, port)
		}
	}

	return net.JoinHostPort(defaultHost, defaultPort)
}
