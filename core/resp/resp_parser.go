package resp

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

type respParser struct {
	buf *bytes.Buffer
}

func NewRESPParser(data []byte) *respParser {
	bufCopy := make([]byte, len(data))
	copy(bufCopy, data)

	return &respParser{
		buf: bytes.NewBuffer(bufCopy),
	}
}

// Parse takes in a RESP based message and returns
// back a 2D byte string that returns back the
// RESP message without the metadata.
func (r *respParser) Parse() ([][]byte, error) {
	return r.parse()
}

func (r *respParser) parse() ([][]byte, error) {
	reqs := make([][]byte, 0)

	for {
		b, err := r.buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("unable to read byte from buffer")
		}

		switch b {
		case '+':
			reqs = append(reqs, r.readSimpleString())
		case '$':
			reqs = append(reqs, r.readBulkString())
		case ':':
			reqs = append(reqs, r.readInteger())
		case '*':
			reqs = append(reqs, r.readArray())
		default:
			return nil, fmt.Errorf("command sent with a protocol that the server doesn't understand")
		}

		if err != nil {
			return nil, err
		}
	}

	for _, req := range reqs {
		fmt.Println(string(req))
	}

	return reqs, nil
}

// Takes in strings of the format +{COMMAND}\r\n
// and returns back a byte array containing only the {COMMAND}
func (r *respParser) readSimpleString() []byte {
	return r.readUntilDelimiter()
}

// Takes in strings of the format ${LENGTH}\r\n{COMMAND}\r\n
// and returns back a byte array containing only the {COMMAND}
func (r *respParser) readBulkString() []byte {
	b := r.readUntilDelimiter()
	expectedStrLen, err := strconv.Atoi(string(b))
	if err != nil {
		panic("error converting string to int")
	}

	b = r.buf.Next(expectedStrLen)
	r.buf.Next(2)
	return b
}

// Takes in strings of the format :{INTEGER}\r\n
// and returns back a byte array containing only the {INTEGER}
func (r *respParser) readInteger() []byte {
	return r.readUntilDelimiter()
}

// Takes in strings of the format *{ARR_LEN}\r\n${LENGTH}\r\n{COMMAND}\r\n${LENGTH}\r\n{COMMAND}\r\n
// and returns back a 2D byte array with each row representing one command.
//
// Note that an array can contain other types but I have limited it to
// bulk strings for simplicity.
func (r *respParser) readArray() []byte {
	b := r.readUntilDelimiter()
	expectedArrayLen, err := strconv.Atoi(string(b))
	if err != nil {
		panic("error converting string to int")
	}

	reqs := make([]byte, 0)
	for i := 0; i < expectedArrayLen; i++ {
		// We need to skip `$` character associated
		// with the bulk string syntax
		r.buf.Next(1)
		reqs = append(reqs, r.readBulkString()...)

		// We add a space as a delimiter
		reqs = append(reqs, []byte(" ")...)
	}
	return reqs
}

// readUntilDelimiter reads and returns the bytes until
// the next seen delimiter. It also takes care of skipping
// the delimiters and so the readers can safely continue
// parsing the further set of bytes.
func (r *respParser) readUntilDelimiter() []byte {
	idx := bytes.Index(r.buf.Bytes(), []byte{'\r', '\n'})
	if idx == -1 {
		// This should never happen
		panic("unable to identify delimiter for reading a string")
	}

	b := r.buf.Next(idx)
	r.skipDelimiter()
	return b
}

// skipDelimiter shifts the buffer by a length of 2
// to ignore the delimiters
func (r *respParser) skipDelimiter() {
	r.buf.Next(2)
}
