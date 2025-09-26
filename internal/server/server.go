package server

import (
	"HTTP_FROM_TCP/internal/headers"
	"HTTP_FROM_TCP/internal/request"
	"HTTP_FROM_TCP/internal/response"
	"fmt"
	"io"
	"log"
	"net"
)

// type serverState string

type Server struct {
	handler Handler
	closed  bool
}

type HandlerError struct {
	Status  response.StatusCode
	Message string
}

type Handler func(w *response.Writer, req *request.Request, defaultHeaders *headers.Headers) error

// this is where we will do something with the connection
func (s *Server) handleConn(conn io.ReadWriteCloser) {
	defer conn.Close()

	writer := response.NewWriter(conn)

	defaultHeaders := response.GetDefaultHeaders(0)
	request, err := request.RequestFromReader(conn)

	if err != nil {
		writer.WriteStatusLine(response.StatusBadRequest)
		writer.WriteHeaders(&defaultHeaders)
		return
	}

	err = s.handler(writer, request, &defaultHeaders)
	if err != nil {
		log.Fatal(err)
	}
}

// this runServer runs in a infinite loop
// the listener is a blocking call and blocks till a client attempts to make a connection with the server
// once the client connects a conn is created and to handle the connection a new go routine handleConn is spawned
// this is done so that each connection gets its own go routine which handles the connection
func (s *Server) runServer(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if s.closed {
			return
		}

		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

// the Serve creates listener which listens to client connections
// spawns a go routine runServer and passes the listener and the server struct to it
// this is done so that the server runs in the background and the main Serve function exists
// and we can continue to do our work
func Serve(port int, handler Handler) (*Server, error) {
	server := &Server{
		closed:  false,
		handler: handler,
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	go server.runServer(listener)
	return server, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}
