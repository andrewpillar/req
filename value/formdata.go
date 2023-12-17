package value

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/andrewpillar/req/syntax"
)

// FormData is the value for data encoded as multipart/form-data. This holds
// the Content-Type of the data, which would be set in a request header, and the
// data itself.
type FormData struct {
	ContentType string
	Data        *bytes.Reader
}

// ToFormData attempts to type assert the given value to FormData.
func ToFormData(v Value) (*FormData, error) {
	f, ok := v.(*FormData)

	if !ok {
		return nil, typeError(v.valueType(), formDataType)
	}
	return f, nil
}

// String formats the FormData value to a string. The formatted string will
// contain the pointer at which FormData exists.
func (f *FormData) String() string {
	return fmt.Sprintf("FormData<addr=%p>", f)
}

// Sprint returns the verbatim string representation of the FormData.
func (f *FormData) Sprint() string {
	if f.Data == nil {
		return ""
	}

	b, err := io.ReadAll(f.Data)

	if err != nil {
		return ""
	}

	if _, err := f.Data.Seek(0, io.SeekStart); err != nil {
		return ""
	}

	return string(b)
}

// Select will return the value of the field with the given name.
func (f *FormData) Select(val Value) (Value, error) {
	name, err := ToName(val)

	if err != nil {
		return nil, err
	}

	switch name.Value {
	case "Content-Type":
		return String{
			Value: f.ContentType,
		}, nil
	case "Data":
		return NewStream(BufferStream(f.Data)), nil
	default:
		return nil, errors.New("type " + val.valueType().String() + " has no field " + name.Value)
	}
}

func (f *FormData) valueType() valueType {
	return formDataType
}

func (f *FormData) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, formDataType)
}
