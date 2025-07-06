package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/AkuPython/Learn-the-HTTP-Protocol/internal/request"
	"github.com/AkuPython/Learn-the-HTTP-Protocol/internal/response"
)

type Server struct {
	handler  Handler
	listener net.Listener
	closed   atomic.Bool
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, h Handler) (*Server, error) {
	l, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		handler:  h,
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
	w := response.Writer{W: conn}
	req, err := request.RequestFromReader(conn)
	if err != nil {
		w.WriteStatusLine(response.StatusBadRequest)
		body := []byte(fmt.Sprintf("Error parsing request: %v", err))
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
		return
	}

	s.handler(&w, req)

	return
}
