package game

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	net2 "github.com/jadefish/avatar/pkg/net"
)

// Server is a capable of communicating with clients currently in the game
// world.
type Server struct {
	stopping bool
	addr     string
	listener net.Listener
	clients  []*client
}

// NewServer creates a new game server.
func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

type client struct {
	conn net.Conn
	ok   bool
}

func (c *client) disconnect() error {
	err := c.conn.Close()
	c.ok = false
	return err
}

func dumpData(first byte, data []byte, n int) {
	log.Printf("cmd 0x%X, %d bytes:\n%s\n", first, n, hex.Dump(data[:n]))
}

// Start the server, allowing it to accept and process incoming client
// connections.
func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.addr)

	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s.listener = l

	log.Println("game: listening on", s.Address())

	for {
		if s.stopping {
			break
		}

		conn, err := s.listener.Accept()

		if err != nil {
			return fmt.Errorf("accept: %w", err)
		}

		client := &client{
			conn: conn,
			ok:   true,
		}
		s.clients = append(s.clients, client)

		errs := make(chan error)

		go func() {
			for {
				if !client.ok {
					break
				}

				// 1. recv unencrypted 4-byte seed from login server's
				// 0x8C Connect to Game Server packet
				// 2. recv 0x91 Game Server Login (encrypted, 65 bytes)

				buf := make([]byte, net2.MaxPacketSize)
				n, err := client.conn.Read(buf)
				errs <- err

				if n == 0 {
					errs <- errors.New("no data!")
					errs <- client.disconnect()
					break
				}

				first := buf[0]
				dumpData(first, buf, n)

				// TODO: could recv n > 0 bytes and an EOF. need to process n
				// bytes before processing EOF. "EOF disconnect" should probably
				// only happen iff n == 0 and err == EOF.
				if err == io.EOF {
					errs <- client.disconnect()
					break
				}
			}
		}()

		go func() {
			for err := range errs {
				if err != nil {
					log.Println(err)
				}
			}
		}()
	}

	return nil
}

// Stop the server, disconnecting connected clients and preventing new
// connections from being accepted.
func (s *Server) Stop() error {
	log.Println("Stopping server...")

	s.stopping = true

	for _, client := range s.clients {
		err := client.disconnect()

		if err != nil {
			log.Printf("server stop: %s\n", err)
		}
	}

	return s.listener.Close()
}

// Address retrieves the server's address.
func (s Server) Address() string {
	return s.listener.Addr().String()
}
