package value

import (
	"time"

	"github.com/andrewpillar/req/syntax"
)

type Time struct {
	Value time.Time
}

func (t Time) String() string {
	return t.Value.Format(time.RFC1123)
}

func (t Time) Sprint() string {
	return t.String()
}

func (t Time) valueType() valueType {
	return timeType
}

func (t Time) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, timeType)
}
