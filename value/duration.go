package value

import (
	"time"

	"github.com/andrewpillar/req/syntax"
)

type Duration struct {
	Value time.Duration
}

func ToDuration(v Value) (Duration, error) {
	d, ok := v.(Duration)

	if !ok {
		return Duration{}, typeError(v.valueType(), durationType)
	}
	return d, nil
}

func (d Duration) String() string {
	return d.Value.String()
}

func (d Duration) Sprint() string {
	return d.String()
}

func (d Duration) valueType() valueType {
	return durationType
}

func (d Duration) cmp(op syntax.Op, b Value) (Value, error) {
	a := Int{Value: int64(d.Value)}

	return a.cmp(op, b)
}
