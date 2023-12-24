package server

import (
	"fmt"
	"syscall"

	"github.com/ganeshrockz/go-redis/protocol"
)

type Server struct {
	SignalCh chan struct{}
	ErrCh    chan error

	protocol protocol.Protocol
}

func New() *Server {
	return &Server{
		SignalCh: make(chan struct{}),
		ErrCh:    make(chan error),
		protocol: protocol.NewMsgLenProtocol(),
	}
}

func (s *Server) RunServer() {
	fmt.Println("Starting server...")
	// Create a socket for IPV4 and TCP connections
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil || fd == 0 {
		panic("can't create socket")
	}

	sockAddr := &syscall.SockaddrInet4{
		Port: 1234,
		Addr: [4]byte{0},
	}

	err = syscall.Bind(fd, sockAddr)
	if err != nil {
		panic(fmt.Sprintf("unable to bind to address %s", err.Error()))
	}

	err = syscall.Listen(fd, syscall.SOMAXCONN)
	if err != nil {
		panic("unable to listen on address")
	}

	close(s.SignalCh)

	for {
		connFd, _, err := syscall.Accept(fd)
		if err != nil {
			s.ErrCh <- fmt.Errorf("error accepting request %w", err)
			continue
		}

		if connFd < 0 {
			s.ErrCh <- fmt.Errorf("fd cannot be less than zero for a new connection")
			continue
		}

		if err = s.handleConnection(connFd); err != nil {
			s.ErrCh <- fmt.Errorf("error handling connection %w", err)
		}

		if err = syscall.Close(connFd); err != nil {
			s.ErrCh <- fmt.Errorf("error closing connection %w", err)
		}
	}
}

func (s *Server) handleConnection(fd int) error {
	for {
		if err := s.serverSingleRequest(fd); err != nil {
			return err
		}
	}
}

func (s *Server) serverSingleRequest(fd int) error {
	data, err := s.protocol.Read(fd)
	if err != nil {
		return fmt.Errorf("reading from file descriptor %w", err)
	}

	fmt.Printf("client says: %s\n", string(data))

	// Respond over the same connection
	if err := s.protocol.Write(fd, "world"); err != nil {
		return fmt.Errorf("error writing back to client %w", err)
	}

	return nil
}
