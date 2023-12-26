package protocol

import (
	"encoding/binary"
	"fmt"
	"syscall"
)

const (
	BufferSize = 4096

	// Size occupied by the length integer of a message.
	// This will be prepended to all the messages.
	messageLengthBytes = 4
)

type messageLengthProtocol struct{}

// This protocol reads and writes messages to the fd in the following format
//
//			|----------|---------|-------|-------|--------|--------|--------|
//			|   len    |   str1  |  len  |  str2 | .......|  len   |   strn |
//			|----------|---------|-------|-------|--------|--------|--------|
//
//	  Here len indicates the length of the string that follows it. Each
//	  message contains the length of the string (representing the first 4 bytes)
//	  followed by the variable length string itself.
func NewMsgLenProtocol() Protocol {
	return &messageLengthProtocol{}
}

func (mlp *messageLengthProtocol) Read(fd int) ([]byte, error) {
	data, err := readAll(fd, messageLengthBytes)
	if err != nil {
		return nil, fmt.Errorf("reading first 4 bytes %w", err)
	}

	msgLen := binary.LittleEndian.Uint32(data)
	if msgLen > BufferSize {
		return nil, fmt.Errorf("reading message. message too long")
	}

	// Read actual message
	data, err = readAll(fd, int(msgLen))
	if err != nil {
		return nil, fmt.Errorf("reading message %s", err.Error())
	}

	return data, nil
}

func (mlp *messageLengthProtocol) Write(fd int, msg string) error {
	lengthBytes := make([]byte, 4)
	data := make([]byte, len(lengthBytes)+BufferSize)

	binary.LittleEndian.PutUint32(data[:4], uint32(len(msg)))
	copy(data[4:], []byte(msg))

	if err := writeAll(fd, data, 4+len(msg)); err != nil {
		return fmt.Errorf("writing message %w", err)
	}

	return nil
}

func readAll(fd int, n int) ([]byte, error) {
	data := make([]byte, 0, n)
	for {
		if n <= 0 {
			break
		}

		temp := make([]byte, n)
		readBytes, err := syscall.Read(fd, temp)
		if err != nil {
			// This indicates that the fd is not ready
			// to be read
			if err == syscall.EAGAIN {
				continue
			}
			return nil, fmt.Errorf("reading file descriptor %w", err)
		}

		if readBytes <= 0 {
			return nil, fmt.Errorf("read error")
		}

		if readBytes > n {
			return nil, fmt.Errorf("read more than what we can hold")
		}

		data = append(data, temp[:readBytes]...)
		n = n - readBytes
	}

	return data, nil
}

func writeAll(fd int, data []byte, n int) error {
	for {
		if n <= 0 {
			break
		}

		temp := make([]byte, n)
		copy(temp, data)

		wroteBytes, err := syscall.Write(fd, temp)
		if err != nil {
			// This indicates that the fd is not ready
			// to be read
			return fmt.Errorf("writing to file descriptor %w", err)
		}

		if wroteBytes <= 0 {
			return fmt.Errorf("write error")
		}

		if wroteBytes > n {
			return fmt.Errorf("wrote more than what the buffer can hold")
		}

		n = n - wroteBytes
		data = data[wroteBytes:]
	}

	return nil
}
