package value

import (
	"encoding/json"

	"github.com/andrewpillar/req/syntax"
)

type Zero struct{}

func (z Zero) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}

func (z Zero) String() string {
	return ""
}

func (z Zero) Sprint() string {
	return ""
}

func (z Zero) valueType() valueType {
	return zeroType
}

func (z Zero) cmp(op syntax.Op, b Value) (Value, error) {
	return b.cmp(op, z)
}
