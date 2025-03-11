package server

import (
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/GhostVox/httptcp/internal/request"
	"github.com/GhostVox/httptcp/internal/response"
)

type Handler func(w response.Writer, req *request.Request)

type HandlerError struct {
	Message    string
	StatusCode response.StatusCode
}

func (he HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.StatusCode)
	messageBytes := []byte(he.Message)
	headers := response.GetDefaultHeaders(len(messageBytes))
	response.WriteHeaders(w, headers)
	w.Write(messageBytes)
}

type Server struct {
	port    int
	server  net.Listener
	handler Handler
	closed  atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {

	tcpListener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	server := &Server{
		port:    port,
		server:  tcpListener,
		handler: handler,
		closed:  atomic.Bool{},
	}
	go server.listen()
	return server, nil

}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.server.Close()

}

func (s *Server) listen() {
	for {
		conn, err := s.server.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Fatal("Failed to accept connection")
		}
		go s.Handle(conn)
	}
}

func (s *Server) Handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.BadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	writer := response.NewResponse(conn)
	s.handler(writer, req)

	return
}
