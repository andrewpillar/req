package value

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/andrewpillar/req/syntax"
)

type FormData struct {
	ContentType string
	Data        *bytes.Reader
}

func ToFormData(v Value) (*FormData, error) {
	f, ok := v.(*FormData)

	if !ok {
		return nil, typeError(v.valueType(), formDataType)
	}
	return f, nil
}

func (f *FormData) String() string {
	return fmt.Sprintf("FormData<addr=%p>", f)
}

func (f *FormData) Sprint() string {
	b, err := io.ReadAll(f.Data)

	if err != nil {
		return ""
	}

	f.Data.Seek(0, io.SeekStart)
	return string(b)
}

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
		return &memStream{
			SectionReader: io.NewSectionReader(f.Data, 0, int64(f.Data.Len())),
		}, nil
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
