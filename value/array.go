package value

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"hash/fnv"

	"github.com/andrewpillar/req/syntax"
)

// Array holds a list of values, and an underlying hash of each of the items
// in the array. This hash is used to perform in operations on the array.
type Array struct {
	curr  int
	atEOF bool

	set   map[uint32]struct{}
	Items []Value
}

// NewArray returns an array with the given list of values. A type check is
// performed on the array to ensure that they all contain the same type.
func NewArray(items []Value) (*Array, error) {
	if len(items) > 1 {
		var typ valueType

		for _, it := range items {
			if it.valueType() != typ {
				if typ == valueType(0) {
					typ = it.valueType()
					continue
				}
				return nil, errors.New("array can only contain type " + typ.String())
			}
		}
	}

	v := &Array{
		set:   make(map[uint32]struct{}),
		Items: items,
	}

	v.hashItems()

	return v, nil
}

// hashItems hashes the string representation of each item in the array and
// stores it in a table. This hash is performed using 32-bit FNV-1a hash.
func (a *Array) hashItems() {
	for _, it := range a.Items {
		h := fnv.New32a()
		h.Write([]byte(it.String()))

		a.set[h.Sum32()] = struct{}{}
	}
}

func (a *Array) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Items)
}

// Has returns whether or not the given value exists in the array.
func (a *Array) Has(v Value) bool {
	if a.set == nil {
		return false
	}

	h := fnv.New32a()
	h.Write([]byte(v.String()))

	_, ok := a.set[h.Sum32()]
	return ok
}

// Get returns the value at the given index in the underlying array. If the
// given value cannot be used as an Int then an error is returned. If the
// index is out of bounds then the Zero value is returned.
func (a *Array) Get(v Value) (Value, error) {
	i64, err := ToInt(v)

	if err != nil {
		return nil, err
	}

	i := int(i64.Value)

	if i < 0 || i > len(a.Items)-1 {
		return Zero{}, nil
	}
	return a.Items[i], nil
}

// Set sets the value at the given key with the given value.
func (a *Array) Set(strict bool, key, val Value) error {
	if _, ok := key.(*Array); ok {
		if len(a.Items) > 0 {
			if err := CompareType(val, a.Items[0]); err != nil {
				return err
			}
		}
		a.Items = append(a.Items, val)
		return nil
	}

	i64, err := ToInt(key)

	if err != nil {
		return err
	}

	i := int(i64.Value)

	if i < 0 || i > len(a.Items)-1 {
		return errors.New("assignment out of bounds")
	}

	if err := CompareType(val, a.Items[i]); err != nil {
		return err
	}

	a.Items[i] = val
	return nil
}

func (a *Array) Next() (Value, Value, error) {
	if a.curr > len(a.Items) - 1 {
		// Reset the current for the next time the value is iterated over.
		a.curr = 0

		return nil, nil, io.EOF
	}

	it := a.Items[a.curr]
	i := a.curr

	a.curr++

	return Int{Value: int64(i)}, it, nil
}

// String formats the array into a string. Each item in the array is space
// separated and wrapped in [ ]. The underlying items in the array will have
// the String method called on them for formatting.
func (a *Array) String() string {
	var buf bytes.Buffer

	buf.WriteByte('[')

	end := len(a.Items) - 1

	for i, it := range a.Items {
		buf.WriteString(it.String())

		if i != end {
			buf.WriteByte(' ')
		}
	}

	buf.WriteByte(']')
	return buf.String()
}

// Sprint is similar to String, the only difference being the Sprint method is
// called on each items in the array for formatting.
func (a *Array) Sprint() string {
	var buf bytes.Buffer

	buf.WriteByte('[')

	end := len(a.Items) - 1

	for i, it := range a.Items {
		buf.WriteString(it.Sprint())

		if i != end {
			buf.WriteByte(' ')
		}
	}

	buf.WriteByte(']')
	return buf.String()
}

func (a *Array) valueType() valueType {
	return arrayType
}

func (a *Array) cmp(op syntax.Op, b Value) (Value, error) {
	typ := b.valueType()

	if typ != arrayType {
		if typ != zeroType {
			return nil, compareError(op, a, b)
		}
	}

	other := b.(*Array)

	ans := false
	invert := false

	switch op {
	case syntax.NeqOp:
		invert = true
		fallthrough
	case syntax.EqOp:
		if typ == zeroType {
			ans = len(a.Items) == 0
			break
		}

		if len(a.Items) != len(other.Items) {
			ans = false
			break
		}

		if len(a.Items) == 0 && len(other.Items) == 0 {
			ans = true
			break
		}

		for i, it := range a.Items {
			val, err := it.cmp(op, other.Items[i])

			if err != nil {
				ans = false
				break
			}
			ans = Truthy(val)
		}
	default:
		return nil, opError(op, arrayType)
	}

	if invert {
		ans = !ans
	}
	return Bool{Value: ans}, nil
}
