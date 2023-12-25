package connection

import "fmt"

type ConnectionRegistry interface {
	Add(conn *Connection) error
	Retrieve(fd int) (*Connection, error)
}

type registry struct {
	conns map[int]*Connection
}

func NewConnRegistry() ConnectionRegistry {
	return &registry{
		conns: make(map[int]*Connection),
	}
}

func (r *registry) Add(conn *Connection) error {
	if _, ok := r.conns[conn.fd]; ok {
		return fmt.Errorf("connection already present in the registry")
	}

	r.conns[conn.fd] = conn
	return nil
}

func (r *registry) Retrieve(fd int) (*Connection, error) {
	conn, ok := r.conns[fd]
	if !ok {
		return nil, fmt.Errorf("connection not present in the registry")
	}

	return conn, nil
}
