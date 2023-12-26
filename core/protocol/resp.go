package protocol

import (
	"bytes"
	"fmt"
	"syscall"
)

const (
	maxBufferSize = 50 * 1024
	bufferSize    = 512
)

type respProtocol struct {
}

func NewRESPProtocol() Protocol {
	return &respProtocol{}
}

func (s *respProtocol) Read(fd int) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	for {
		tempBuf := make([]byte, bufferSize)
		readBytes, err := syscall.Read(fd, tempBuf)
		if err != nil {
			if err == syscall.EAGAIN {
				continue
			}
			return nil, fmt.Errorf("error reading from fd %w", err)
		}
		if readBytes <= 0 {
			// Read is complete
			break
		}
		tempBuf = tempBuf[:readBytes]
		buf.Write(tempBuf)

		if bytes.Contains(tempBuf, []byte{'\r', '\n'}) {
			break
		}

		if buf.Len() > maxBufferSize {
			return nil, fmt.Errorf("message too long. capacity %d bytes, actual %d bytes", maxBufferSize, buf.Len())
		}
	}

	return buf.Bytes(), nil
}

func (s *respProtocol) Write(fd int, resp string) error {
	n := len(resp)
	data := []byte(resp)
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
