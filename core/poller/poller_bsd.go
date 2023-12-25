//go:build darwin || netbsd || freebsd || openbsd || dragonfly
// +build darwin netbsd freebsd openbsd dragonfly

package poller

import (
	"fmt"
	"syscall"

	"github.com/ganeshrockz/go-redis/core/events"
)

type poller struct {
	kq int
}

func NewPoller() Poller {
	return &poller{}
}

func (p *poller) Setup(fd int) (int, error) {
	var err error
	p.kq, err = syscall.Kqueue()
	if err != nil {
		return -1, fmt.Errorf("unable to create a kqueue: %w", err)
	}

	kEvent := events.NewKernelEvent(syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
	})

	registered, err := kEvent.Register(p.kq)
	if err != nil || !registered {
		return -1, fmt.Errorf("unable to register change event")
	}

	return p.kq, nil
}

func (p *poller) Poll() ([]*events.KernelEvent, error) {
	newEvents := make([]syscall.Kevent_t, 10)
	numNewEvents, err := syscall.Kevent(p.kq, nil, newEvents, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading new events")
	}

	if numNewEvents == 0 {
		return nil, nil
	}

	kernelEvents := make([]*events.KernelEvent, 0)
	for _, event := range newEvents {
		kernelEvents = append(kernelEvents, events.NewKernelEvent(event))
	}

	return kernelEvents, nil
}
