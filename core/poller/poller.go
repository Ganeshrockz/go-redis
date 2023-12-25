package poller

import "github.com/ganeshrockz/go-redis/core/events"

// Poller defines the interface for polling events
// out of the kernel
type Poller interface {
	// Setup accepts a file descriptor of a socket and sets up the necessary
	// infra to listen to events against the same.
	Setup(int) (int, error)

	// Poll polls the Kqueue and accumulates the new set of events.
	Poll() ([]*events.KernelEvent, error)
}
