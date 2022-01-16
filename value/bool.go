package value

import (
	"encoding/json"

	"github.com/andrewpillar/req/syntax"
)

// Bool is the value for boolean types.
type Bool struct {
	Value bool
}

// ToBool attempts to type assert the given value to a bool.
func ToBool(v Value) (Bool, error) {
	b, ok := v.(Bool)

	if !ok {
		return Bool{}, typeError(v.valueType(), boolType)
	}
	return b, nil
}

func (b Bool) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Value)
}

func (b Bool) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

func (b Bool) Sprint() string {
	return b.String()
}

func (b Bool) valueType() valueType {
	return boolType
}

func (bl Bool) cmp(op syntax.Op, b Value) (Value, error) {
	typ := b.valueType()

	if typ != boolType {
		if typ != zeroType {
			return nil, compareError(op, bl, b)
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
			ans = !bl.Value
			break
		}
		ans = bl.Value == b.(Bool).Value
	default:
		return nil, opError(op, boolType)
	}

	if invert {
		ans = !ans
	}
	return Bool{Value: ans}, nil
}
