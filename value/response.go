package value

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/andrewpillar/req/syntax"
)

type Response struct {
	*http.Response
}

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
	case "Header":
		pairs := make(map[string]Value)

		for k, v := range r.Header {
			pairs[k] = String{Value: v[0]}
		}
		return Object{Pairs: pairs}, nil
	case "Body":
		if r.Body == nil {
			return &stream{}, nil
		}

		rc, rc2 := copyrc(r.Body)
		r.Body = rc

		b, _ := io.ReadAll(rc2)

		return &stream{r: bytes.NewReader(b)}, nil
	default:
		return nil, errors.New("type " + val.valueType().String() + " has no field " + name.Value)
	}
}

func (r Response) String() string {
	return fmt.Sprintf("Response<addr=%p>", r.Response)
}

func (r Response) Sprint() string {
	buf := bytes.NewBufferString(r.Proto + " " + r.Status + "\n")

	r.Header.Write(buf)

	if r.Body != nil {
		buf.WriteString("\n")

		rc, rc2 := copyrc(r.Body)

		r.Body = rc
		io.Copy(buf, rc2)
	}
	return buf.String()
}

func (r Response) valueType() valueType {
	return responseType
}

func (r Response) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, responseType)
}
