package value

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
)

type valueType uint

//go:generate stringer -type valueType -linecomment
const (
	stringType   valueType = iota + 1 // string
	intType                           // int
	boolType                          // bool
	arrayType                         // array
	objectType                        // object
	fileType                          // file
	requestType                       // request
	responseType                      // response
	streamType                        // stream
	nameType                          // name
	zeroType                          // zero
)

// Type is a convenience function that returns the type of the given value.
func Type(v Value) string {
	return v.valueType().String()
}

func typeError(typ1, typ2 valueType) error {
	return fmt.Errorf("cannot use %s as %s\n", typ1, typ2)
}

// Stream represents a stream of data that can be read. This would either be a
// Request/Response body or a File.
type Stream interface {
	io.ReadSeeker
	io.Closer
}

func ToStream(v Value) (Stream, error) {
	s, ok := v.(Stream)

	if !ok {
		return nil, typeError(v.valueType(), streamType)
	}
	return s, nil
}

type Value interface {
	// String formats the Value into a string. The returned string is suitable
	// for display in a REPL. For example, strings are quoted.
	String() string

	// Sprint formats the Value into a string. This differs from String, in
	// that the returned string may not be suitable for display in a REPL.
	// For example, strings are not quoted, and the entire contents of Streams
	// are returned.
	Sprint() string

	valueType() valueType
}

// Index represents a Value that can be indexed, such as an Object or an Array.
type Index interface {
	// Has checks to see if the given Value exists in the underlying index.
	Has(Value) bool

	Get(Value) (Value, error)
}

func ToIndex(v Value) (Index, error) {
	i, ok := v.(Index)

	if !ok {
		return nil, errors.New("type " + v.valueType().String() + " does not support indexing")
	}
	return i, nil
}

type Selector interface {
	Select(Value) (Value, error)
}

func ToSelector(v Value) (Selector, error) {
	s, ok := v.(Selector)

	if !ok {
		return nil, typeError(v.valueType(), streamType)
	}
	return s, nil
}

type String struct {
	Value string
}

func ToString(v Value) (String, error) {
	s, ok := v.(String)

	if !ok {
		return String{}, typeError(v.valueType(), stringType)
	}
	return s, nil
}

func (s String) String() string {
	return fmt.Sprintf("%q", s.Value)
}

func (s String) Sprint() string {
	return s.Value
}

func (s String) valueType() valueType {
	return stringType
}

type Int struct {
	Value int64
}

func ToInt(v Value) (Int, error) {
	i, ok := v.(Int)

	if !ok {
		return Int{}, typeError(v.valueType(), intType)
	}
	return i, nil
}

func (i Int) String() string {
	return strconv.FormatInt(i.Value, 10)
}

func (i Int) Sprint() string {
	return i.String()
}

func (i Int) valueType() valueType {
	return intType
}

type Bool struct {
	Value bool
}

func ToBool(v Value) (Bool, error) {
	b, ok := v.(Bool)

	if !ok {
		return Bool{}, typeError(v.valueType(), boolType)
	}
	return b, nil
}

func (b Bool) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

func (b Bool) Sprint() string {
	return b.String()
}

func (b Bool) valueType() valueType {
	return boolType
}

type Array struct {
	set   map[uint32]struct{}
	Items []Value
}

func NewArray(items []Value) (*Array, error) {
	if len(items) > 1 {
		typ1 := items[0].valueType()
		typ2 := items[1].valueType()

		if typ1 != typ2 {
			return nil, errors.New("array can only contain type " + typ1.String())
		}
	}

	v := &Array{
		set:   make(map[uint32]struct{}),
		Items: items,
	}

	for _, it := range v.Items {
		h := fnv.New32a()
		h.Write([]byte(it.String()))

		v.set[h.Sum32()] = struct{}{}
	}
	return v, nil
}

func (a *Array) Has(v Value) bool {
	if a.set == nil {
		return false
	}

	h := fnv.New32a()
	h.Write([]byte(v.String()))

	_, ok := a.set[h.Sum32()]
	return ok
}

func (a *Array) Get(v Value) (Value, error) {
	i64, err := ToInt(v)

	if err != nil {
		return nil, err
	}

	i := int(i64.Value)

	if i < 0 || i > len(a.Items)-1 {
		return Zero{}, nil
	}
	return a.Items[i], nil
}

func (a *Array) String() string {
	var buf bytes.Buffer

	buf.WriteByte('[')

	end := len(a.Items) - 1

	for i, it := range a.Items {
		buf.WriteString(it.String())

		if i != end {
			buf.WriteByte(' ')
		}
	}

	buf.WriteByte(']')
	return buf.String()
}

func (a *Array) Sprint() string {
	var buf bytes.Buffer

	buf.WriteByte('[')

	end := len(a.Items) - 1

	for i, it := range a.Items {
		buf.WriteString(it.Sprint())

		if i != end {
			buf.WriteByte(' ')
		}
	}

	buf.WriteByte(']')
	return buf.String()
}

func (a *Array) valueType() valueType {
	return arrayType
}

type Object struct {
	Pairs map[string]Value
}

func ToObject(v Value) (Object, error) {
	o, ok := v.(Object)

	if !ok {
		return Object{}, typeError(v.valueType(), objectType)
	}
	return o, nil
}

func (o Object) Has(v Value) bool {
	if o.Pairs == nil {
		return false
	}

	str, ok := v.(String)

	if !ok {
		return false
	}

	_, ok = o.Pairs[str.Value]
	return ok
}

func (o Object) Get(v Value) (Value, error) {
	str, err := ToString(v)

	if err != nil {
		return nil, err
	}

	val, ok := o.Pairs[str.Value]

	if !ok {
		return Zero{}, nil
	}
	return val, nil
}

func (o Object) String() string {
	var buf bytes.Buffer

	buf.WriteByte('{')

	end := len(o.Pairs) - 1
	i := 0

	for k, v := range o.Pairs {
		buf.WriteString(k + ":")
		buf.WriteString(v.String())

		if i != end {
			buf.WriteByte(' ')
		}
	}

	buf.WriteByte('}')
	return buf.String()
}

func (o Object) Sprint() string {
	var buf bytes.Buffer

	buf.WriteByte('{')

	order := make([]string, 0, len(o.Pairs))

	for k := range o.Pairs {
		order = append(order, k)
	}

	sort.Strings(order)

	end := len(o.Pairs) - 1

	for i, k := range order {
		buf.WriteString(k+":"+o.Pairs[k].Sprint())

		if i != end {
			buf.WriteByte(' ')
		}
	}

	buf.WriteByte('}')
	return buf.String()
}

func (o Object) valueType() valueType {
	return objectType
}

type File struct {
	*os.File
}

func (f File) Read(p []byte) (int, error) {
	if f.File == nil {
		return 0, nil
	}
	return f.File.Read(p)
}

func (f File) Seek(offset int64, whence int) (int64, error) {
	if f.File == nil {
		return 0, nil
	}
	return f.File.Seek(offset, whence)
}

func (f File) Close() error {
	if f.File == nil {
		return nil
	}
	return f.File.Close()
}

func (f File) String() string {
	name := ""

	if f.File != nil {
		name = f.Name()
	}
	return fmt.Sprintf("File<addr=%p, name=%q>", f.File, name)
}

func (f File) Sprint() string {
	b, err := io.ReadAll(f.File)

	if err != nil {
		return ""
	}

	f.Seek(0, io.SeekStart)
	return string(b)
}

func (f File) valueType() valueType {
	return fileType
}

type stream struct {
	closed bool
	r      *bytes.Reader
}

var errStreamClosed = errors.New("stream closed")

func (s *stream) Read(p []byte) (int, error) {
	if s.closed {
		return 0, errStreamClosed
	}
	return s.r.Read(p)
}

func (s *stream) Seek(offset int64, whence int) (int64, error) {
	if s.closed {
		return 0, errStreamClosed
	}
	return s.r.Seek(offset, whence)
}

func (s *stream) Close() error {
	if s.closed {
		return errStreamClosed
	}

	s.closed = true
	return nil
}

func (s *stream) String() string {
	return fmt.Sprintf("Stream<addr=%p>", s.r)
}

func (s *stream) Sprint() string {
	b, err := io.ReadAll(s.r)

	if err != nil {
		return ""
	}

	s.r.Seek(0, io.SeekStart)
	return string(b)
}

func (s *stream) valueType() valueType {
	return streamType
}

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

		return &stream{r: bytes.NewReader(b)}, nil
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

type Name struct {
	Value string
}

func ToName(v Value) (Name, error) {
	n, ok := v.(Name)

	if !ok {
		return Name{}, typeError(v.valueType(), nameType)
	}
	return n, nil
}

func (n Name) String() string {
	return n.Value
}

func (n Name) Sprint() string {
	return n.String()
}

func (n Name) valueType() valueType {
	return nameType
}

type Zero struct{}

func (z Zero) String() string {
	return ""
}

func (z Zero) Sprint() string {
	return ""
}

func (z Zero) valueType() valueType {
	return zeroType
}
