package value

import (
	"bytes"
	"sort"

	"github.com/andrewpillar/req/syntax"
)

type Object struct {
	Pairs map[string]Value
}

func ToObject(v Value) (Object, error) {
	o, ok := v.(Object)

	if !ok {
		return Object{}, typeError(v.valueType(), objectType)
	}
	return o, nil
}

func (o Object) Has(v Value) bool {
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

func (o Object) Get(v Value) (Value, error) {
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

func (o Object) String() string {
	var buf bytes.Buffer

	buf.WriteByte('{')

	end := len(o.Pairs) - 1
	i := 0

	for k, v := range o.Pairs {
		buf.WriteString(k + ":")
		buf.WriteString(v.String())

		if i != end {
			buf.WriteByte(' ')
		}
	}

	buf.WriteByte('}')
	return buf.String()
}

func (o Object) Sprint() string {
	var buf bytes.Buffer

	buf.WriteByte('{')

	order := make([]string, 0, len(o.Pairs))

	for k := range o.Pairs {
		order = append(order, k)
	}

	sort.Strings(order)

	end := len(o.Pairs) - 1

	for i, k := range order {
		buf.WriteString(k+":"+o.Pairs[k].Sprint())

		if i != end {
			buf.WriteByte(' ')
		}
	}

	buf.WriteByte('}')
	return buf.String()
}

func (o Object) valueType() valueType {
	return objectType
}

func (o Object) cmp(op syntax.Op, b Value) (Value, error) {
	typ := b.valueType()

	if typ != objectType {
		if typ != zeroType {
			return nil, compareError(op, o, b)
		}
	}

	other := b.(Object)

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
