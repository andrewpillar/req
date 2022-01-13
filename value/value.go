package value

import (
	"errors"
	"fmt"
	"io"

	"github.com/andrewpillar/req/syntax"
)

type valueType uint

//go:generate stringer -type valueType -linecomment
const (
	stringType   valueType = iota + 1 // string
	intType                           // int
	boolType                          // bool
	arrayType                         // array
	objectType                        // object
	fileType                          // file
	requestType                       // request
	responseType                      // response
	streamType                        // stream
	nameType                          // name
	zeroType                          // zero
)

// Type is a convenience function that returns the type of the given value.
func Type(v Value) string {
	return v.valueType().String()
}

func typeError(typ1, typ2 valueType) error {
	return fmt.Errorf("cannot use %s as %s", typ1, typ2)
}

// Stream represents a stream of data that can be read. This would either be a
// Request/Response body or a File.
type Stream interface {
	io.ReadSeeker
	io.Closer
}

func ToStream(v Value) (Stream, error) {
	s, ok := v.(Stream)

	if !ok {
		return nil, typeError(v.valueType(), streamType)
	}
	return s, nil
}

type Value interface {
	// String formats the Value into a string. The returned string is suitable
	// for display in a REPL. For example, strings are quoted.
	String() string

	// Sprint formats the Value into a string. This differs from String, in
	// that the returned string may not be suitable for display in a REPL.
	// For example, strings are not quoted, and the entire contents of Streams
	// are returned.
	Sprint() string

	valueType() valueType

	cmp(syntax.Op, Value) (Value, error)
}

// Index represents a Value that can be indexed, such as an Object or an Array.
type Index interface {
	// Has checks to see if the given Value exists in the underlying index.
	Has(Value) bool

	Get(Value) (Value, error)
}

func ToIndex(v Value) (Index, error) {
	i, ok := v.(Index)

	if !ok {
		return nil, errors.New("type " + v.valueType().String() + " does not support indexing")
	}
	return i, nil
}

type Selector interface {
	Select(Value) (Value, error)
}

func ToSelector(v Value) (Selector, error) {
	s, ok := v.(Selector)

	if !ok {
		return nil, typeError(v.valueType(), streamType)
	}
	return s, nil
}