package client

import (
	"fmt"
	"syscall"

	"github.com/ganeshrockz/go-redis/core/protocol"
)

type Client struct {
	protocol protocol.Protocol
}

func New() *Client {
	return &Client{
		protocol: protocol.NewMsgLenProtocol(),
	}
}

func (c *Client) Run() {
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

	defer func() {
		if err = syscall.Close(fd); err != nil {
			panic(fmt.Sprintf("error closing connection %s\n", err.Error()))
		}
	}()

	if err = c.req(fd, "hello1"); err != nil {
		panic(fmt.Sprintf("talking to server %s", err.Error()))
	}

	if err = c.req(fd, "hello2"); err != nil {
		panic(fmt.Sprintf("talking to server %s", err.Error()))
	}

	if err = c.req(fd, "hello3"); err != nil {
		panic(fmt.Sprintf("talking to server %s", err.Error()))
	}

	if err = c.res(fd); err != nil {
		panic(fmt.Sprintf("reading from server %s", err.Error()))
	}

	if err = c.res(fd); err != nil {
		panic(fmt.Sprintf("reading from server %s", err.Error()))
	}

	if err = c.res(fd); err != nil {
		panic(fmt.Sprintf("reading from server %s", err.Error()))
	}
}

func (c *Client) query(fd int, resp string) error {
	if err := c.protocol.Write(fd, resp); err != nil {
		return fmt.Errorf("writing to file descriptor %w", err)
	}

	data, err := c.protocol.Read(fd)
	if err != nil {
		return fmt.Errorf("reading from file descriptor %w", err)
	}

	fmt.Printf("server says: %s\n", string(data))
	return nil
}

func (c *Client) req(fd int, resp string) error {
	if err := c.protocol.Write(fd, resp); err != nil {
		return fmt.Errorf("writing to file descriptor %w", err)
	}
	return nil
}

func (c *Client) res(fd int) error {
	data, err := c.protocol.Read(fd)
	if err != nil {
		return fmt.Errorf("reading from file descriptor %w", err)
	}

	fmt.Printf("server says: %s\n", string(data))
	return nil
}
