package connection

import (
	"fmt"

	"github.com/ganeshrockz/go-redis/core/protocol"
)

type Connection struct {
	fd       int
	protocol protocol.Protocol
}

func NewConnection(fd int) *Connection {
	return &Connection{
		fd:       fd,
		protocol: protocol.NewMsgLenProtocol(),
	}
}

func (c *Connection) Handle() error {
	data, err := c.protocol.Read(c.fd)
	if err != nil {
		return fmt.Errorf("reading from file descriptor %w", err)
	}

	fmt.Printf("client says: %s\n", string(data))

	// Respond over the same connection
	if err := c.protocol.Write(c.fd, "world"); err != nil {
		return fmt.Errorf("error writing back to client %w", err)
	}

	return nil
}
