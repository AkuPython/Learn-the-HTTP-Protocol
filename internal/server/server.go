package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/AkuPython/Learn-the-HTTP-Protocol/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	l, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		listener: l,
	}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()

}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("error: could not accept connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	size := 0
	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		log.Printf("error: could not write status-line")
	}
	dh := response.GetDefaultHeaders(size)
	err = response.WriteHeaders(conn, dh)
	if err != nil {
		log.Printf("error: could not write headers")
	}
	return
}
