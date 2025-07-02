package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
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
	// defer l.Close()
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
			log.Fatal(err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	placeholder := make([]byte, 8096)
	_, err := conn.Read(placeholder)
	if err != nil {
		log.Fatal(err)
	}
	response := fmt.Sprint(
		"HTTP/1.1 200 OK\r\n",
		"Content-Type: text/plain\r\n\r\n",
		"Hello World!\r\n",
	)

	conn.Write([]byte(response))
}
