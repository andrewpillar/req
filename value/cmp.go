package value

import "github.com/andrewpillar/req/syntax"

type compareFunc func(a, b Value) (Value, error)

var cmptab = map[syntax.Op]compareFunc{
	syntax.EqOp:  cmpEq,
	syntax.NeqOp: cmpNeq,
	syntax.LtOp:  cmpLt,
	syntax.LeqOp: cmpLeq,
	syntax.GtOp:  cmpGt,
	syntax.GeqOp: cmpGeq,
	syntax.InOp:  cmpIn,
	syntax.AndOp: cmpAnd,
	syntax.OrOp:  cmpOr,
}

func cmpEq(a, b Value) (Value, error) {

	return nil, nil
}

func cmpNeq(a, b Value) (Value, error) {

	return nil, nil
}

func cmpLt(a, b Value) (Value, error) {

	return nil, nil
}

func cmpLeq(a, b Value) (Value, error) {

	return nil, nil
}

func cmpGt(a, b Value) (Value, error) {

	return nil, nil
}

func cmpGeq(a, b Value) (Value, error) {

	return nil, nil
}

func cmpIn(a, b Value) (Value, error) {

	return nil, nil
}

func cmpAnd(a, b Value) (Value, error) {

	return nil, nil
}

func cmpOr(a, b Value) (Value, error) {
	return nil, nil
}

type CmpError struct {
	Op   syntax.Op
	A, B Value
}

func (e *CmpError) Error() string {
	return "type mismatch for comparison: " +
		e.A.valueType().String() + " " +
		e.Op.String() + " " +
		e.B.valueType().String()
}

func Compare(a Value, op syntax.Op, b Value) (Value, error) {
	if a.valueType() != b.valueType() {
		return nil, &CmpError{
			Op: op, A: a, B: b,
		}
	}

	cmp, ok := cmptab[op]

	if !ok {
		panic("invalid comparator: " + op.String())
	}
	return cmp(a, b)
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
