package eval

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sort"
)

type valueType uint

//go:generate stringer -type valueType -linecomment
const (
	stringType valueType = iota + 1 // string
	intType                         // int
	boolType                        // bool
	arrayType                       // array
	objectType                      // object
	fileType                        // file
	respType                        // resp
)

type Value interface {
	String() string

	type_() valueType
}

type String struct {
	Value string
}

func (s String) String() string { return s.Value }

func (s String) type_() valueType { return stringType }

type Int struct {
	Value int64
}

func (i Int) String() string { return strconv.FormatInt(i.Value, 10) }

func (i Int) type_() valueType { return intType }

type Bool struct {
	Value bool
}

func (b Bool) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

func (b Bool) type_() valueType { return boolType }

type Array struct {
	Items []Value
}

func (a Array) String() string {
	var buf bytes.Buffer

	buf.WriteString("[")

	end := len(a.Items) - 1

	for i, it := range a.Items {
		buf.WriteString(it.String())

		if i != end {
			buf.WriteString(", ")
		}
	}
	return buf.String()
}

func (a Array) type_() valueType { return arrayType }

type Key struct {
	Name string
}

type Object struct {
	Pairs map[Key]Value
}

func (o Object) String() string {
	var buf bytes.Buffer

	order := make([]string, 0, len(o.Pairs))
	pairs := make(map[string]string)

	for k, v := range o.Pairs {
		buf.WriteString(`"`+k.Name+`": `+v.String())

		order = append(order, k.Name)
		pairs[k.Name] = buf.String()

		buf.Reset()
	}

	buf.Reset()
	sort.Strings(order)

	buf.WriteString("{")

	end := len(order) - 1

	for i, name := range order {
		buf.WriteString(pairs[name])

		if i != end {
			buf.WriteString(", ")
		}
	}

	buf.WriteString("}")
	return buf.String()
}

func (o Object) type_() valueType { return objectType }

type File struct {
	*os.File
}

func (f File) String() string {
	return fmt.Sprintf("File<addr=%p, name=%q>", f.File, f.Name())
}

func (f File) type_() valueType { return fileType }

type Resp struct {
	*http.Response
}

func copyrc(rc io.ReadCloser) (io.ReadCloser, io.ReadCloser) {
	var buf bytes.Buffer

	buf.ReadFrom(rc)
	rc.Close()

	return io.NopCloser(&buf), io.NopCloser(bytes.NewBuffer(buf.Bytes()))
}

func (r Resp) String() string {
	var buf bytes.Buffer

	buf.WriteString(r.Proto + " ")
	buf.WriteString(r.Status + "\n")

	r.Header.Write(&buf)
	buf.WriteString("\n\n")

	rc, rc2 := copyrc(r.Body)

	buf.ReadFrom(rc)

	r.Body = rc2

	return buf.String()
}

func (r Resp) type_() valueType { return respType }

type Yield struct {
	Value Value
}

func (y Yield) String() string { return "yield" }

func (y Yield) type_() valueType { return valueType(0) }
