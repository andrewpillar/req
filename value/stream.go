package value

import (
	"errors"
	"fmt"
	"io"

	"github.com/andrewpillar/req/syntax"
)

var errStreamClosed = errors.New("stream closed")

type baseStream struct {
	closed bool
}

func stringStream(s Stream) string {
	return fmt.Sprintf("Stream<addr=%p>", s)
}

func sprintStream(s Stream) string {
	b, err := io.ReadAll(s)

	if err != nil {
		return ""
	}

	s.Seek(0, io.SeekStart)
	return string(b)
}

func (s *baseStream) valueType() valueType {
	return streamType
}

func (s *baseStream) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, streamType)
}

func (s *baseStream) Close() error {
	if s.closed {
		return errStreamClosed
	}
	s.closed = true
	return nil
}

type memStream struct {
	baseStream
	*io.SectionReader
}

func (m *memStream) String() string {
	return stringStream(m)
}

func (m *memStream) Sprint() string {
	return sprintStream(m)
}

func (m *memStream) Read(p []byte) (int, error) {
	if m.closed {
		return 0, errStreamClosed
	}
	return m.SectionReader.Read(p)
}

func (m *memStream) ReadAt(p []byte, off int64) (int, error) {
	if m.closed {
		return 0, errStreamClosed
	}
	return m.SectionReader.ReadAt(p, off)
}

func (m *memStream) Seek(offset int64, whence int) (int64, error) {
	if m.closed {
		return 0, errStreamClosed
	}
	return m.SectionReader.Seek(offset, whence)
}

func (m *memStream) Close() error {
	return m.baseStream.Close()
}

type valStream struct {
	baseStream
	Stream
}

func NewStream(s Stream) Value {
	if v, ok := s.(Value); ok {
		return v
	}
	return &valStream{Stream: s}
}

func (s *valStream) String() string {
	return stringStream(s.Stream)
}

func (s *valStream) Sprint() string {
	return sprintStream(s.Stream)
}

func (s *valStream) Close() error {
	return s.baseStream.Close()
}
