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
		clients:   make([]Client, 0, clientSliceSize),
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

	done := make(chan bool)
	errs := make(chan error)

	defer close(done)
	defer close(errs)

	for {
		client, err := s.accept()

		if err != nil {
			return err
		}

		go s.processClient(client, done, errs)

		select {
		case <-done:
			s.addClient(*client)
		case e := <-errs:
			log.Println(e)
			client.Disconnect(0x00) // TODO
		}
	}
}

// Stop the server, disconnecting connected clients and preventing new
// connections from being accepted.
func (s *Server) Stop() error {
	for _, c := range s.clients {
		err := c.Disconnect(0x00) // TODO

		if err != nil {
			log.Println(errors.Wrap(err, "server stop"))
		}

		s.removeClient(&c)
	}

	return s.listener.Close()
}

// AccountService returns the server's account service.
func (s Server) AccountService() avatar.AccountService {
	return s.accounts
}

// PasswordService returns the server's password service.
func (s Server) PasswordService() avatar.PasswordService {
	return s.passwords
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

	client, err := NewClient(conn)

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

	// Verify credentials:
	if !s.passwords.VerifyPassword(result.Password, []byte(account.Password)) {
		errs <- errors.Wrap(avatar.ErrInvalidCredentials, "process client")
		return
	}

	done <- true
}

// addClient adds the provided client to the server's list of clients.
func (s *Server) addClient(client Client) {
	s.clients = append(s.clients, client)
}

// removeClient removes the provided client from the server's list of
// clients, returning true if a client was removed.
// The removed client (if any) is not disconnected.
func (s *Server) removeClient(client *Client) bool {
	for i, c := range s.clients {
		if c == *client {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			return true
		}
	}

	return false
}

func getAddr() string {
	if val, ok := os.LookupEnv("LOGIN_ADDR"); ok {
		if host, port, err := net.SplitHostPort(val); err == nil {
			return net.JoinHostPort(host, port)
		}
	}

	return net.JoinHostPort(defaultHost, defaultPort)
}
