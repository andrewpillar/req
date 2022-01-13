package value

import (
	"bytes"
	"encoding/json"
	"errors"
	"hash/fnv"

	"github.com/andrewpillar/req/syntax"
)

type Array struct {
	set   map[uint32]struct{}
	Items []Value
}

func NewArray(items []Value) (*Array, error) {
	if len(items) > 1 {
		typ1 := items[0].valueType()
		typ2 := items[1].valueType()

		if typ1 != typ2 {
			return nil, errors.New("array can only contain type " + typ1.String())
		}
	}

	v := &Array{
		set:   make(map[uint32]struct{}),
		Items: items,
	}

	v.hashItems()

	return v, nil
}

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

func (a *Array) Has(v Value) bool {
	if a.set == nil {
		return false
	}

	h := fnv.New32a()
	h.Write([]byte(v.String()))

	_, ok := a.set[h.Sum32()]
	return ok
}

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
