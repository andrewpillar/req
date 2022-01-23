package value

import (
	"encoding/json"
	"strconv"

	"github.com/andrewpillar/req/syntax"
)

type Float struct {
	Value float64
}

func (f Float) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Value)
}

func (f Float) String() string {
	return strconv.FormatFloat(f.Value, 'f', 2, 64)
}

func (f Float) Sprint() string {
	return f.String()
}

func (f Float) valueType() valueType {
	return floatType
}

func (f Float) cmp(op syntax.Op, b Value) (Value, error) {
	typ := b.valueType()

	if typ != intType {
		if typ != zeroType {
			return nil, compareError(op, f, b)
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
			ans = f.Value == 0
			break
		}
		ans = f.Value == b.(Float).Value
	case syntax.LtOp:
		if typ == zeroType {
			ans = f.Value < 0
			break
		}
		ans = f.Value < b.(Float).Value
	case syntax.LeqOp:
		if typ == zeroType {
			ans = f.Value <= 0
			break
		}
		ans = f.Value <= b.(Float).Value
	case syntax.GtOp:
		if typ == zeroType {
			ans = f.Value > 0
			break
		}
		ans = f.Value > b.(Float).Value
	case syntax.GeqOp:
		if typ == zeroType {
			ans = f.Value >= 0
			break
		}
		ans = f.Value >= b.(Float).Value
	default:
		return nil, opError(op, intType)
	}

	if invert {
		ans = !ans
	}
	return Bool{Value: ans}, nil
}
