package value

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/andrewpillar/req/syntax"
)

// Response is the value for an HTTP response. This holds the underlying handle
// to the response.
type Response struct {
	*http.Response
}

// Select will return the value of the field with the given name.
func (r Response) Select(val Value) (Value, error) {
	name, err := ToName(val)

	if err != nil {
		return nil, err
	}

	switch name.Value {
	case "Status":
		return String{Value: r.Status}, nil
	case "StatusCode":
		return Int{Value: int64(r.StatusCode)}, nil
	case "Cookie":
		cookies := r.Cookies()

		pairs := make(map[string]Value)
		order := make([]string, 0, len(cookies))

		for _, ck := range cookies {
			order = append(order, ck.Name)
			pairs[ck.Name] = Cookie{
				Cookie: ck,
			}
		}

		return &Object{
			Order: order,
			Pairs: pairs,
		}, nil
	case "Header":
		pairs := make(map[string]Value)
		order := make([]string, 0, len(r.Header))

		for k, v := range r.Header {
			order = append(order, k)

			vals := make([]Value, 0, len(v))

			for _, s := range v {
				vals = append(vals, String{Value: s})
			}

			arr, err := NewArray(vals)

			if err != nil {
				return nil, err
			}

			pairs[k] = &Tuple{
				t1: vals[0],
				t2: arr,
			}
		}
		return &Object{
			Order: order,
			Pairs: pairs,
		}, nil
	case "Body":
		if r.Body == nil {
			return &stream{}, nil
		}

		rc, rc2, err := copyrc(r.Body)

		if err != nil {
			return nil, err
		}
		r.Body = rc

		b, _ := io.ReadAll(rc2)

		return NewStream(BufferStream(bytes.NewReader(b))), nil
	default:
		return nil, errors.New("type " + val.valueType().String() + " has no field " + name.Value)
	}
}

// String formats the response to a string. The formatted string will detail the
// pointer at which the underlying response handle exists.
func (r Response) String() string {
	return fmt.Sprintf("Response<addr=%p>", r.Response)
}

// Sprint formats the response into a string. This makes a copy of the response
// body so as to not deplete the original.
func (r Response) Sprint() string {
	if r.Response == nil {
		return ""
	}

	buf := bytes.NewBufferString(r.Proto + " " + r.Status + "\n")

	if err := r.Header.Write(buf); err != nil {
		panic(err)
	}

	if r.Body != nil {
		buf.WriteString("\n")

		rc, rc2, err := copyrc(r.Body)

		if err != nil {
			panic(err)
		}

		r.Body = rc
		if _, err := io.Copy(buf, rc2); err != nil {
			panic(err)
		}
	}
	return buf.String()
}

func (r Response) valueType() valueType {
	return responseType
}

func (r Response) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, responseType)
}
