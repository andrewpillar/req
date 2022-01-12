package value

import (
	"fmt"

	"github.com/andrewpillar/req/syntax"
)

type String struct {
	Value string
}

func ToString(v Value) (String, error) {
	s, ok := v.(String)

	if !ok {
		return String{}, typeError(v.valueType(), stringType)
	}
	return s, nil
}

func (s String) String() string {
	return fmt.Sprintf("%q", s.Value)
}

func (s String) Sprint() string {
	return s.Value
}

func (s String) valueType() valueType {
	return stringType
}

func (s String) cmp(op syntax.Op, b Value) (Value, error) {
	typ := b.valueType()

	if typ != stringType {
		if typ != zeroType {
			return nil, compareError(op, s, b)
		}
	}

	ans := false
	invert := false

	switch op {
	case syntax.NeqOp:
		invert = true
		fallthrough
	case syntax.EqOp:
		if typ == zeroType {
			ans = s.Value == ""
			break
		}
		ans = s.Value == b.(String).Value
	case syntax.LtOp:
		if typ == zeroType {
			ans = s.Value < ""
			break
		}
		ans = s.Value < b.(String).Value
	case syntax.LeqOp:
		if typ == zeroType {
			ans = s.Value <= ""
			break
		}
		ans = s.Value <= b.(String).Value
	case syntax.GtOp:
		if typ == zeroType {
			ans = s.Value > ""
			break
		}
		ans = s.Value > b.(String).Value
	case syntax.GeqOp:
		if typ == zeroType {
			ans = s.Value >= ""
			break
		}
		ans = s.Value >= b.(String).Value
	default:
		return nil, opError(op, stringType)
	}

	if invert {
		ans = !ans
	}
	return Bool{Value: ans}, nil
}
