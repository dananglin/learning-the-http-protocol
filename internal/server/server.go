package server

import (
	"fmt"
	"log/slog"
	"net"
	"sync/atomic"

	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	listener net.Listener
	closed   *atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	address := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("error creating the listener: %w", err)
	}

	closed := atomic.Bool{}
	closed.Store(false)

	server := Server{
		listener: listener,
		closed:   &closed,
		handler:  handler,
	}

	go server.listen()

	slog.Info("HTTP server is now accepting web requests", "address", address)

	return &server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)

	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("error closing the listener: %w", err)
	}

	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				break
			}

			slog.Error("error accepting the connection.", "error", err.Error())

			continue
		}

		slog.Info("Connection accepted.")

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		slog.Error("error parsing the request.", "error", err.Error())

		return
	}

	resp := response.NewWriter(conn)

	s.handler(resp, req)
}
