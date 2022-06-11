package value

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/andrewpillar/req/syntax"
)

type Cookie struct {
	*http.Cookie
}

func (c Cookie) Select(val Value) (Value, error) {
	name, err := ToName(val)

	if err != nil {
		return nil, err
	}

	switch name.Value {
	case "Name":
		return String{Value: c.Name}, nil
	case "Value":
		return String{Value: c.Value}, nil
	case "Path":
		return String{Value: c.Path}, nil
	case "Domain":
		return String{Value: c.Domain}, nil
//	case "Expires":
	case "MaxAge":
		return Int{Value: int64(c.MaxAge)}, nil
	case "Secure":
		return Bool{Value: c.Secure}, nil
	case "HttpOnly":
		return Bool{Value: c.HttpOnly}, nil
//	case "SameSite":
	default:
		return nil, errors.New("type " + val.valueType().String() + " has no field " + name.Value)
	}
}

// String formats the cookie to a string. The formatted string will detail the
// pointer at which the underlying response handle exists.
func (c Cookie) String() string {
	return fmt.Sprintf("Cookie<addr=%p>", c.Cookie)
}

func (c Cookie) Sprint() string {
	return c.Cookie.String()
}

func (c Cookie) valueType() valueType {
	return cookieType
}

func (c Cookie) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, cookieType)
}
