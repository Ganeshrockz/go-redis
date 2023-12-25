package events

import "syscall"

// A wrapper on top of Kevent_t
type KernelEvent struct {
	k_event syscall.Kevent_t
}

func NewKernelEvent(k_event syscall.Kevent_t) *KernelEvent {
	return &KernelEvent{
		k_event: k_event,
	}
}

func (k *KernelEvent) IsCloseEvent() bool {
	return k.k_event.Flags&syscall.EV_EOF != 0
}

func (k *KernelEvent) EventFD() uint64 {
	return k.k_event.Ident
}

func (k *KernelEvent) IsReadEvent() bool {
	return k.k_event.Filter&syscall.EVFILT_READ != 0
}

func (k *KernelEvent) Register(kq int) (bool, error) {
	registered, err := syscall.Kevent(kq, []syscall.Kevent_t{k.k_event}, nil, nil)
	if err != nil {
		return false, err
	}

	return registered != -1, nil
}
