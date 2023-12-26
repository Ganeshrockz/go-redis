package server

import (
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/ganeshrockz/go-redis/core/command"
	"github.com/ganeshrockz/go-redis/core/connection"
	"github.com/ganeshrockz/go-redis/core/events"
	"github.com/ganeshrockz/go-redis/core/poller"
	"github.com/ganeshrockz/go-redis/core/store"
)

var (
	statusNotStarted int64 = 0
	statusWaiting    int64 = 1
	statusBusy       int64 = 2
	statusTerminate  int64 = 3
)

type Server struct {
	SignalCh chan struct{}
	ErrCh    chan error
	wg       *sync.WaitGroup
	sigs     chan os.Signal
	status   int64
}

func New(wg *sync.WaitGroup, sigs chan os.Signal) *Server {
	return &Server{
		SignalCh: make(chan struct{}),
		ErrCh:    make(chan error),
		wg:       wg,
		sigs:     sigs,
		status:   statusNotStarted,
	}
}

func (s *Server) RunServer() {
	go s.waitForSIGTERM()
	defer func() {
		atomic.StoreInt64(&s.status, statusTerminate)
	}()

	fmt.Println("Starting server...")
	// Create a socket for IPV4 and TCP connections
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil || fd == 0 {
		errStr := "can't create socket"
		s.signalServerErr(fmt.Errorf(errStr))
		panic(errStr)
	}

	sockAddr := &syscall.SockaddrInet4{
		Port: 1234,
		Addr: [4]byte{0},
	}

	err = syscall.Bind(fd, sockAddr)
	if err != nil {
		errStr := fmt.Sprintf("unable to bind to address %s", err.Error())
		s.signalServerErr(fmt.Errorf(errStr))
		panic(errStr)
	}

	err = syscall.Listen(fd, syscall.SOMAXCONN)
	if err != nil {
		errStr := "unable to listen on address"
		s.signalServerErr(fmt.Errorf(errStr))
		panic(errStr)
	}

	// Signal that the server has started
	close(s.SignalCh)

	err = syscall.SetNonblock(fd, true)
	if err != nil {
		errStr := "unable to set server socket's fd to non blocking"
		s.signalServerErr(fmt.Errorf(errStr))
		panic(errStr)
	}

	eventPoller := poller.NewPoller()
	kq, err := eventPoller.Setup(fd)
	if err != nil || kq == -1 {
		errStr := "error setting up event poller"
		s.signalServerErr(fmt.Errorf(errStr))
		panic(errStr)
	}

	// Initialize the store
	store := store.NewMapStore()

	// Initialize the connection registry
	connectionPool := connection.NewConnRegistry()

	// Register the commands
	commandRegistry := command.NewRegistry()
	command.RegisterCommands(commandRegistry)

	for atomic.LoadInt64(&s.status) != statusTerminate {
		eventsToHandle, err := eventPoller.Poll()
		if err != nil {
			s.signalServerErr(fmt.Errorf("error polling new events %w", err))
		}
		if eventsToHandle == nil {
			log.Printf("no new events to poll")
			continue
		}

		switch atomic.LoadInt64(&s.status) {
		case int64(statusTerminate):
			return
		case int64(statusWaiting):
			atomic.StoreInt64(&s.status, statusBusy)
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

				if err := connectionPool.Add(connection.NewConnection(connFd, store, commandRegistry)); err != nil {
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
					s.signalServerErr(fmt.Errorf("unable to set client connection's socket fd to non blocking"))
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

			atomic.StoreInt64(&s.status, statusWaiting)
		}
	}
}

func (s *Server) waitForSIGTERM() {
	defer s.wg.Done()
	<-s.sigs

	for atomic.LoadInt64(&s.status) == statusBusy {
	}

	atomic.StoreInt64(&s.status, statusTerminate)

	os.Exit(0)
}

func (s *Server) signalServerErr(err error) {
	s.ErrCh <- err
}
