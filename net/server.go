package net

import (
	"net"
	"os"

	"github.com/jadefish/avatar"
)

// TODO: load me from configuration:
const (
	clientSliceSize = 256
	defaultHost     = "0.0.0.0"
	defaultPort     = "7775"
)

type Server struct {
	clients  []*avatar.Client
	addr     net.Addr
	listener net.Listener
}

func NewServer() *Server {
	return &Server{
		clients: make([]*avatar.Client, clientSliceSize),
	}
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", getAddr())

	if err != nil {
		return err
	}

	s.listener = l
	s.addr = l.Addr()

	return nil
}

func (s *Server) Stop() error {
	return s.listener.Close()
}

func (s *Server) Address() string {
	return s.listener.Addr().String()
}

func (s *Server) Accept() (net.Conn, error) {
	return s.listener.Accept()
}

func getAddr() string {
	if val, ok := os.LookupEnv("LOGIN_ADDR"); ok {
		if host, port, err := net.SplitHostPort(val); err == nil {
			return net.JoinHostPort(host, port)
		}
	}

	return net.JoinHostPort(defaultHost, defaultPort)
}

func (s *Server) AddClient(client *avatar.Client) error {
	s.clients = append(s.clients, client)

	return nil
}
