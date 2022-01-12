package value

import (
	"errors"
	"fmt"

	"github.com/andrewpillar/req/syntax"
)

func cmpIn(a, b Value) (Value, error) {
	index, err := ToIndex(b)

	if err != nil {
		return nil, err
	}

	return Bool{
		Value: index.Has(a),
	}, nil
}

func cmpAnd(a, b Value) (Value, error) {
	if Truthy(a) && Truthy(b) {
		return Bool{Value: true}, nil
	}
	return Bool{}, nil
}

func cmpOr(a, b Value) (Value, error) {
	if Truthy(a) || Truthy(b) {
		return Bool{Value: true}, nil
	}
	return Bool{}, nil
}

func compareError(op syntax.Op, a, b Value) error {
	return fmt.Errorf("invalid operation: %s %s %s (cannot compare %s with %s)", a.String(), op, b.String(), a.valueType(), b.valueType())
}

func opError(op syntax.Op, typ valueType) error {
	return errors.New("cannot perform " + op.String() + " on " + typ.String())
}

func Compare(a Value, op syntax.Op, b Value) (Value, error) {
	set := map[syntax.Op]struct{}{
		syntax.EqOp:  {},
		syntax.NeqOp: {},
		syntax.LtOp:  {},
		syntax.LeqOp: {},
		syntax.GtOp:  {},
		syntax.GeqOp: {},
		syntax.InOp:  {},
		syntax.AndOp: {},
		syntax.OrOp:  {},
	}

	if _, ok := set[op]; !ok {
		panic("invalid comparator: " + op.String())
	}

	switch op {
	case syntax.AndOp:
		return cmpAnd(a, b)
	case syntax.OrOp:
		return cmpOr(a, b)
	case syntax.InOp:
		return cmpIn(a, b)
	default:
		return a.cmp(op, b)
	}
}

func CompareType(a, b Value) error {
	if a.valueType() != b.valueType() {
		return typeError(a.valueType(), b.valueType())
	}
	return nil
}

func Truthy(v Value) bool {
	if b, ok := v.(Bool); ok {
		return b.Value
	}
	return false
}
