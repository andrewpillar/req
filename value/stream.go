package value

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/andrewpillar/req/syntax"
)

type stream struct {
	closed bool
	r      *bytes.Reader
}

var errStreamClosed = errors.New("stream closed")

func (s *stream) Read(p []byte) (int, error) {
	if s.closed {
		return 0, errStreamClosed
	}
	return s.r.Read(p)
}

func (s *stream) Seek(offset int64, whence int) (int64, error) {
	if s.closed {
		return 0, errStreamClosed
	}
	return s.r.Seek(offset, whence)
}

func (s *stream) Close() error {
	if s.closed {
		return errStreamClosed
	}

	s.closed = true
	return nil
}

func (s *stream) String() string {
	return fmt.Sprintf("Stream<addr=%p>", s.r)
}

func (s *stream) Sprint() string {
	b, err := io.ReadAll(s.r)

	if err != nil {
		return ""
	}

	s.r.Seek(0, io.SeekStart)
	return string(b)
}

func (s *stream) valueType() valueType {
	return streamType
}

func (s *stream) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, streamType)
}
