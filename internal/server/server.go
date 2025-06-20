package server

import (
	"fmt"
	"log/slog"
	"net"
	"sync/atomic"

	"http-from-tcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   *atomic.Bool
}

func Serve(port int) (*Server, error) {
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

		s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	headers := response.GetDefaultHeaders(0)

	if err := response.WriteStatusLine(conn, response.StatusCodeOK); err != nil {
		slog.Error("error writing the status line.", "error", err.Error())
	}

	if err := response.WriteHeaders(conn, headers); err != nil {
		slog.Error("error writing the headers.", "error", err.Error())
	}
}
