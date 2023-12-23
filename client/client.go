package client

import (
	"fmt"
	"syscall"
)

const bufferSize = 1024

func RunClient() {
	fmt.Println("Starting Client...")
	// Create a socket for IPV4 and TCP connections
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil || fd == 0 {
		panic("can't create socket")
	}

	addr := &syscall.SockaddrInet4{
		Port: 1234,
		Addr: [4]byte{127, 0, 0, 1},
	}

	err = syscall.Connect(fd, addr)
	if err != nil {
		panic(fmt.Sprintf("unable to connect to socket %s", err.Error()))
	}

	msg := "Hello server"
	n, err := syscall.Write(fd, []byte(msg))
	if err != nil {
		panic(fmt.Sprintf("writing response %s", err.Error()))
	}

	if n < 0 {
		panic("write error")
	}

	resp := make([]byte, bufferSize)
	n, err = syscall.Read(fd, resp)
	if err != nil {
		panic(fmt.Sprintf("reading response %s", err.Error()))
	}

	if n < 0 {
		panic("read error")
	}

	fmt.Printf("Server says %s\n", string(resp))

	if err = syscall.Close(fd); err != nil {
		panic(fmt.Sprintf("error closing connection %s\n", err.Error()))
	}
}
