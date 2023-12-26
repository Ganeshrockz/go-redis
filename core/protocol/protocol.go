package protocol

// Protocol is the interface any individual protocol should implement
type Protocol interface {
	// Read accepts a file descriptor and contains logic to read
	// from the same depending on the protocol
	Read(int) ([]byte, error)

	// Write accepts a file descriptor and an input string and contains logic to write
	// the input string to the same connection.
	Write(int, string) error
}
