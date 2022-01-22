package value

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"

	"github.com/andrewpillar/req/syntax"
)

// Object holds a list of values indexed under a string.
type Object struct {
	curr  int
	atEOF bool

	Order []string // The order in which the keys should be iterated.

	Pairs map[string]Value
}

// ToObjectt attempts to type assert the given value to an object.
func ToObject(v Value) (*Object, error) {
	o, ok := v.(*Object)

	if !ok {
		return nil, typeError(v.valueType(), objectType)
	}
	return o, nil
}

func (o *Object) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Pairs)
}

// Has checks to see if the current object has the current value, if that given
// value is a string.
func (o *Object) Has(v Value) bool {
	if o.Pairs == nil {
		return false
	}

	str, ok := v.(String)

	if !ok {
		return false
	}

	_, ok = o.Pairs[str.Value]
	return ok
}

// Get returns the value at the given index, if that value is a string. If there
// is no value at the given index, then Zero is returned.
func (o *Object) Get(v Value) (Value, error) {
	str, err := ToString(v)

	if err != nil {
		return nil, err
	}

	val, ok := o.Pairs[str.Value]

	if !ok {
		return Zero{}, nil
	}
	return val, nil
}

// Set sets the value at the given key with the given value.
func (o *Object) Set(strict bool, key, val Value) error {
	str, err := ToString(key)

	if err != nil {
		return err
	}

	val0, ok := o.Pairs[str.Value]

	if !ok {
		o.Pairs[str.Value] = val
		return nil
	}

	if strict {
		if err := CompareType(val, val0); err != nil {
			return err
		}
	}

	o.Pairs[str.Value] = val
	return nil
}

func (o *Object) Next() (Value, Value, error) {
	if o.curr > len(o.Order) - 1 {
		// Reset the current for the next time the value is iterated over.
		o.curr = 0

		return nil, nil, io.EOF
	}

	key := o.Order[o.curr]
	o.curr++

	return String{Value:key}, o.Pairs[key], nil
}

// String formats the object into a string. Each key-value pair is space
// spearated and wrapped in ( ). The underlying values in the array will have
// the String method called on them for formatting.
func (o *Object) String() string {
	var buf bytes.Buffer

	buf.WriteByte('(')

	end := len(o.Pairs) - 1
	i := 0

	for k, v := range o.Pairs {
		buf.WriteString(k + ":")
		buf.WriteString(v.String())

		if i != end {
			buf.WriteByte(' ')
		}
		i++
	}

	buf.WriteByte(')')
	return buf.String()
}

// Sprint is similar to String, the only difference being the Sprint method is
// called on each value in the object for formatting.
func (o *Object) Sprint() string {
	var buf bytes.Buffer

	buf.WriteByte('(')

	order := make([]string, 0, len(o.Pairs))

	for k := range o.Pairs {
		order = append(order, k)
	}

	sort.Strings(order)

	end := len(o.Pairs) - 1

	for i, k := range order {
		buf.WriteString(k + ":" + o.Pairs[k].Sprint())

		if i != end {
			buf.WriteByte(' ')
		}
	}

	buf.WriteByte(')')
	return buf.String()
}

func (o *Object) valueType() valueType {
	return objectType
}

func (o *Object) cmp(op syntax.Op, b Value) (Value, error) {
	typ := b.valueType()

	if typ != objectType {
		if typ != zeroType {
			return nil, compareError(op, o, b)
		}
	}

	other, _ := b.(*Object)

	ans := false
	invert := false

	switch op {
	case syntax.NeqOp:
		invert = true
		fallthrough
	case syntax.EqOp:
		if typ == zeroType {
			ans = len(o.Pairs) == 0
			break
		}

		if len(o.Pairs) != len(other.Pairs) {
			ans = false
			break
		}

		for k, v := range o.Pairs {
			it, ok := other.Pairs[k]

			if !ok {
				ans = false
				break
			}

			val, err := v.cmp(op, it)

			if err != nil {
				ans = false
				break
			}
			ans = Truthy(val)
		}
	default:
		return nil, opError(op, objectType)
	}

	if invert {
		ans = !ans
	}
	return Bool{Value: ans}, nil
}
