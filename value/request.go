package value

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/andrewpillar/req/syntax"
)

type Request struct {
	*http.Request
}

func ToRequest(v Value) (Request, error) {
	r, ok := v.(Request)

	if !ok {
		return Request{}, typeError(v.valueType(), requestType)
	}
	return r, nil
}

func (r Request) Select(val Value) (Value, error) {
	name, err := ToName(val)

	if err != nil {
		return nil, err
	}

	switch name.Value {
	case "Method":
		return String{Value: r.Method}, nil
	case "URL":
		return String{Value: r.URL.String()}, nil
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

		return NewStream(BufferStream(bytes.NewReader(b))), nil
	default:
		return nil, errors.New("type " + val.valueType().String() + " has no field " + name.Value)
	}
}

func (r Request) String() string {
	return fmt.Sprintf("Request<addr=%p>", r.Request)
}

func copyrc(rc io.ReadCloser) (io.ReadCloser, io.ReadCloser) {
	var buf bytes.Buffer
	buf.ReadFrom(rc)

	return io.NopCloser(&buf), io.NopCloser(bytes.NewBuffer(buf.Bytes()))
}

func (r Request) Sprint() string {
	if r.Request == nil {
		return ""
	}

	buf := bytes.NewBufferString(r.Method + " " + r.Proto + "\n")

	r.Header.Write(buf)

	if r.Body != nil {
		buf.WriteString("\n")

		rc, rc2 := copyrc(r.Body)

		r.Body = rc
		io.Copy(buf, rc2)
	}
	return buf.String()
}

func (r Request) valueType() valueType {
	return requestType
}

func (r Request) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, requestType)
}
