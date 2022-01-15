package value

import (
	"fmt"
	"io"
	"os"

	"github.com/andrewpillar/req/syntax"
)

type File struct {
	*os.File
}

func ToFile(v Value) (File, error) {
	f, ok := v.(File)

	if !ok {
		return File{}, typeError(v.valueType(), fileType)
	}
	return f, nil
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

func (f File) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, fileType)
}
