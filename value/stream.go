package value

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/andrewpillar/req/syntax"
)

type stream struct {
	Stream
}

// NewStream turns the given stream into a value.
func NewStream(s Stream) Value {
	return &stream{
		Stream: s,
	}
}

func (s stream) String() string {
	return fmt.Sprintf("Stream<addr=%p>", s.Stream)
}

func (s stream) Sprint() string {
	b, err := io.ReadAll(s)

	if err != nil {
		return ""
	}

	s.Seek(0, io.SeekStart)
	return string(b)
}

func (s stream) valueType() valueType {
	return streamType
}

func (s stream) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, streamType)
}

type sectionReadCloser struct {
	*io.SectionReader
	closed bool
}

var errClosed = errors.New("stream closed")

// BufferStream returns a new stream for reading data from a location in memory.
func BufferStream(r *bytes.Reader) Stream {
	return &sectionReadCloser{
		SectionReader: io.NewSectionReader(r, 0, int64(r.Len())),
	}
}

func (r *sectionReadCloser) Close() error {
	if r.closed {
		return errClosed
	}
	r.closed = false
	return nil
}
