package value

import "github.com/andrewpillar/req/syntax"

type Bool struct {
	Value bool
}

func ToBool(v Value) (Bool, error) {
	b, ok := v.(Bool)

	if !ok {
		return Bool{}, typeError(v.valueType(), boolType)
	}
	return b, nil
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
