package server

import (
	"fmt"
	"log"
	"syscall"

	"github.com/ganeshrockz/go-redis/core/connection"
	"github.com/ganeshrockz/go-redis/core/events"
	"github.com/ganeshrockz/go-redis/core/poller"
)

type Server struct {
	SignalCh chan struct{}
	ErrCh    chan error
}

func New() *Server {
	return &Server{
		SignalCh: make(chan struct{}),
		ErrCh:    make(chan error),
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

	// Signal that the server has started
	close(s.SignalCh)

	err = syscall.SetNonblock(fd, true)
	if err != nil {
		panic("unable to set server socket's fd to non blocking")
	}

	eventPoller := poller.NewPoller()
	kq, err := eventPoller.Setup(fd)
	if err != nil || kq == -1 {
		panic("error setting up event poller")
	}

	connectionPool := connection.NewConnRegistry()

	for {
		eventsToHandle, err := eventPoller.Poll()
		if err != nil {
			s.signalServerErr(fmt.Errorf("error polling new events %w", err))
		}
		if eventsToHandle == nil {
			log.Printf("no new events to poll")
			continue
		}

		for _, event := range eventsToHandle {
			if event.IsCloseEvent() {
				if err := syscall.Close(int(event.EventFD())); err != nil {
					s.signalServerErr(fmt.Errorf("error closing connection with fd %d: %w", event.EventFD(), err))
				}
			} else if event.EventFD() == uint64(fd) {
				// Handling new connection
				connFd, _, err := syscall.Accept(fd)
				if err != nil {
					s.signalServerErr(fmt.Errorf("error accepting request %w", err))
					continue
				}

				if connFd < 0 {
					s.signalServerErr(fmt.Errorf("fd cannot be less than zero for a new connection"))
					continue
				}

				if err := connectionPool.Add(connection.NewConnection(connFd)); err != nil {
					s.signalServerErr(err)
					continue
				}

				socketEvent := events.NewKernelEvent(syscall.Kevent_t{
					Ident:  uint64(connFd),
					Filter: syscall.EVFILT_READ,
					Flags:  syscall.EV_ADD,
				})
				registered, err := socketEvent.Register(kq)
				if err != nil || !registered {
					s.signalServerErr(fmt.Errorf("unable to register incoming connection's kevent"))
					continue
				}

				err = syscall.SetNonblock(connFd, true)
				if err != nil {
					panic("unable to set client connection's socket fd to non blocking")
				}
			} else if event.IsReadEvent() {
				conn, err := connectionPool.Retrieve(int(event.EventFD()))
				if err != nil {
					s.signalServerErr(err)
					continue
				}

				if err = conn.Handle(); err != nil {
					s.signalServerErr(fmt.Errorf("error handling connection %w", err))
				}
			}
		}
	}
}

func (s *Server) signalServerErr(err error) {
	s.ErrCh <- err
}
