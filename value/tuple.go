package value

import "github.com/andrewpillar/req/syntax"

type Tuple struct {
	T1, T2 Value
}

func (t *Tuple) String() string {
	return t.T1.String()
}

func (t *Tuple) Sprint() string {
	return t.T1.Sprint()
}

func (t *Tuple) valueType() valueType {
	return tupleType
}

func (t *Tuple) cmp(op syntax.Op, b Value) (Value, error) {
	for _, v := range []Value{t.T1, t.T2} {
		ans, _ := v.cmp(op, b)

		if b, ok := ans.(Bool); ok && b.Value {
			return ans, nil
		}
	}
	return nil, opError(op, tupleType)
}
