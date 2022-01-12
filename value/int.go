package value

import (
	"strconv"

	"github.com/andrewpillar/req/syntax"
)

type Int struct {
	Value int64
}

func ToInt(v Value) (Int, error) {
	i, ok := v.(Int)

	if !ok {
		return Int{}, typeError(v.valueType(), intType)
	}
	return i, nil
}

func (i Int) String() string {
	return strconv.FormatInt(i.Value, 10)
}

func (i Int) Sprint() string {
	return i.String()
}

func (i Int) valueType() valueType {
	return intType
}

func (i Int) cmp(op syntax.Op, b Value) (Value, error) {
	typ := b.valueType()

	if typ != intType {
		if typ != zeroType {
			return nil, compareError(op, i, b)
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
			ans = i.Value == 0
			break
		}
		ans = i.Value == b.(Int).Value
	case syntax.LtOp:
		if typ == zeroType {
			ans = i.Value < 0
			break
		}
		ans = i.Value < b.(Int).Value
	case syntax.LeqOp:
		if typ == zeroType {
			ans = i.Value <= 0
			break
		}
		ans = i.Value <= b.(Int).Value
	case syntax.GtOp:
		if typ == zeroType {
			ans = i.Value > 0
			break
		}
		ans = i.Value > b.(Int).Value
	case syntax.GeqOp:
		if typ == zeroType {
			ans = i.Value >= 0
			break
		}
		ans = i.Value >= b.(Int).Value
	default:
		return nil, opError(op, intType)
	}

	if invert {
		ans = !ans
	}
	return Bool{Value: ans}, nil
}
