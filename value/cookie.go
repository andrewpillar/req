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

func ToCookie(v Value) (Cookie, error) {
	c, ok := v.(Cookie)

	if !ok {
		return Cookie{}, typeError(v.valueType(), cookieType)
	}
	return c, nil
}

func (c Cookie) Set(name string, val Value) error {
	setstring := func(p *string) func(v Value) error {
		return func(v Value) error {
			str, err := ToString(v)

			if err != nil {
				return err
			}
			*p = str.Value
			return nil
		}
	}

	setbool := func(p *bool) func(v Value) error {
		return func(v Value) error {
			b, err := ToBool(v)

			if err != nil {
				return err
			}
			*p = b.Value
			return nil
		}
	}

	fieldtab := map[string]func(v Value) error {
		"Name":   setstring(&c.Cookie.Name),
		"Value":  setstring(&c.Cookie.Value),
		"Path":   setstring(&c.Cookie.Path),
		"Domain": setstring(&c.Cookie.Domain),
		"MaxAge": func(v Value) error {
			d, err := ToDuration(v)

			if err != nil {
				return err
			}
			c.Cookie.MaxAge = int(d.Value.Seconds())
			return nil
		},
		"Secure":   setbool(&c.Cookie.Secure),
		"HttpOnly": setbool(&c.Cookie.HttpOnly),
	}

	set, ok := fieldtab[name]

	if !ok {
		return errors.New("unexpected cookie field: " + name)
	}

	if err := set(val); err != nil {
		return errors.New("field error " + name + ": " + err.Error())
	}
	return nil
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
	case "Expires":
		return Time{Value: c.Expires}, nil
	case "MaxAge":
		return Int{Value: int64(c.MaxAge)}, nil
	case "Secure":
		return Bool{Value: c.Secure}, nil
	case "HttpOnly":
		return Bool{Value: c.HttpOnly}, nil
	case "SameSite":
		tab := map[http.SameSite]string{
			http.SameSiteLaxMode:    "Lax",
			http.SameSiteStrictMode: "Strict",
			http.SameSiteNoneMode:   "None",
		}

		if v, ok := tab[c.SameSite]; ok {
			return String{Value: v}, nil
		}
		return Zero{}, nil
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
