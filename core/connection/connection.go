package connection

import (
	"fmt"
	"strings"

	"github.com/ganeshrockz/go-redis/core/command"
	"github.com/ganeshrockz/go-redis/core/protocol"
	"github.com/ganeshrockz/go-redis/core/resp"
	"github.com/ganeshrockz/go-redis/core/store"
)

type Connection struct {
	fd       int
	protocol protocol.Protocol

	store           store.Store
	commandRegistry command.CommandRegistry
}

func NewConnection(fd int, store store.Store, cr command.CommandRegistry) *Connection {
	return &Connection{
		fd:              fd,
		protocol:        protocol.NewRESPProtocol(),
		store:           store,
		commandRegistry: cr,
	}
}

func (c *Connection) Handle() error {
	data, err := c.protocol.Read(c.fd)
	if err != nil {
		return fmt.Errorf("reading from file descriptor %w", err)
	}

	var respStr string
	parser := resp.NewRESPParser(data)
	toProcess, err := parser.Parse()
	if err != nil {
		// Write back to client
		respStr = resp.Encode(fmt.Errorf("parse error %w", err))
		goto write
	}

	respStr = fmt.Sprintf("*%d\r\n", len(toProcess))
	for _, p := range toProcess {
		commandStr := strings.TrimSpace(string(p))
		commandArr := strings.Split(commandStr, " ")
		reg, err := c.commandRegistry.Retrieve(commandArr[0])
		if err != nil {
			respStr = resp.Encode(fmt.Errorf("error executing command %w", err))
			goto write
		}

		if len(commandArr) == 1 {
			// TODO: Handle command arrs of length 1
			continue
		}
		if err = reg.Validate(commandArr[1:]); err != nil {
			respStr = resp.Encode(fmt.Errorf("validation error %w", err))
			goto write
		}

		result, err := reg.Execute(commandArr[1:], c.store)
		if err != nil {
			respStr = resp.Encode(err)
			goto write
		}

		respStr += string(result)
	}

write:
	// Respond over the same connection
	if err := c.protocol.Write(c.fd, respStr); err != nil {
		return fmt.Errorf("error writing back to client %w", err)
	}

	return nil
}
