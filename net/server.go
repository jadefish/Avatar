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
	processCapacity = 5
)

// Server is a capable of accepting and processing clients over a network.
type Server struct {
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

type Result struct {
	Client *Client
	Err    error
}

func (r Result) OK() bool {
	return r.Err == nil
}

func acceptor(id int, server *Server, jobs chan<- *Client) {
	log.Printf("Producer %d: waiting for connection\n", id)

	client, err := server.accept()
	log.Printf("Producer %d: accepted connection %p\n", id, client)

	if err != nil {
		log.Printf("Producer %d: error: %s", id, err.Error())
		return
	}

	log.Printf("Producer %d: sending client %p through jobs channel\n", id, client)
	jobs <- client
}

func processor(id int, server *Server, jobs <-chan *Client, results chan<- Result) {
	log.Printf("Worker %d: waiting for job\n", id)

	for c := range jobs {
		log.Printf("Worker %d: got job for client %p\n", id, c)

		result := Result{
			Client: c,
		}
		err := server.processClient(c)

		if err != nil {
			result.Err = err
		}

		log.Printf("Worker %d: job's done.\n\tResult: %+v\n", id, result)

		results <- result
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

	jobs := make(chan *Client, processCapacity)
	results := make(chan Result, processCapacity)

	// Start client acceptors, which produce client-wrapped connections:
	for w := 0; w < processCapacity; w++ {
		go acceptor(w+1, s, jobs)
	}

	// Start client consumers, which process clients:
	for w := 0; w < processCapacity; w++ {
		go processor(w+1, s, jobs, results)
	}

	for {
		select {
		case r := <-results:
			log.Printf("Got result: %+v\n", r)

			if r.OK() {
				s.addClient(*r.Client)
			} else {
				r.Client.Disconnect(byte(avatar.LoginRejectionInvalidCredentials))
			}
		}
	}
}

// Stop the server, disconnecting connected clients and preventing new
// connections from being accepted.
func (s *Server) Stop() error {
	log.Println("Stopping server...")

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

func (s *Server) processClient(c *Client) error {
	err := c.Connect()

	if err != nil {
		return errors.Wrap(err, "process client")
	}

	result, err := c.Authenticate()

	if err != nil {
		return errors.Wrap(err, "process client")
	}

	// Find account:
	account, err := s.accounts.GetAccountByName(result.AccountName)

	if err != nil {
		return errors.Wrap(err, "process client")
	}

	log.Println("found account:", account)

	// Verify credentials:
	if !s.passwords.VerifyPassword(result.Password, []byte(account.PasswordHash)) {
		return errors.Wrap(avatar.ErrInvalidCredentials, "process client")
	}

	err = c.LogIn()

	if err != nil {
		return errors.Wrap(err, "process client")
	}

	shards, err := s.shards.All()

	if err != nil {
		return errors.Wrap(err, "get shard list")
	}

	err = c.ReceiveShardList(shards)

	if err != nil {
		return errors.Wrap(err, "receive shard list")
	}

	return nil
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
