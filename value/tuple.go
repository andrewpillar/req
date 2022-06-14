package value

import (
	"fmt"

	"github.com/andrewpillar/req/syntax"
)

// Tuple holds two different values that allows for the typle to be used as
// either value.
type Tuple struct {
	t1, t2 Value
}

func (t *Tuple) String() string {
	return fmt.Sprintf("Tuple<%s, %s>", t.t1.valueType(), t.t2.valueType())
}

func (t *Tuple) Sprint() string {
	return t.t1.Sprint()
}

func (t *Tuple) valueType() valueType {
	return tupleType
}

func (t *Tuple) cmp(op syntax.Op, b Value) (Value, error) {
	for _, v := range []Value{t.t1, t.t2} {
		ans, _ := v.cmp(op, b)

		if b, ok := ans.(Bool); ok && b.Value {
			return ans, nil
		}
	}
	return nil, opError(op, tupleType)
}
