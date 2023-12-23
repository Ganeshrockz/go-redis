package server

import (
	"fmt"
	"syscall"
)

const bufferSize = 1024

type Server struct {
	SignalCh chan struct{}
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
			fmt.Printf("error accepting request %s\n", err.Error())
			continue
		}

		if connFd < 0 {
			fmt.Printf("fd cannot be less than zero for a new connection\n")
			continue
		}

		if err = handleConnection(connFd); err != nil {
			fmt.Printf("error handling connection %s", err.Error())
		}

		if err = syscall.Close(connFd); err != nil {
			fmt.Printf("error closing connection %s\n", err.Error())
		}
	}
}

func handleConnection(fd int) error {
	data := make([]byte, bufferSize)
	n, err := syscall.Read(fd, data)
	if err != nil {
		return fmt.Errorf("reading file descriptor %w", err)
	}

	if n < 0 {
		return fmt.Errorf("read error")
	}

	fmt.Printf("client says: %s\n", string(data))

	data = []byte("response from server")
	n, err = syscall.Write(fd, data)
	if err != nil {
		return fmt.Errorf("writing response %w", err)
	}

	if n < 0 {
		return fmt.Errorf("write error")
	}

	return nil
}
