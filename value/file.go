package value

import (
	"fmt"
	"io"
	"os"

	"github.com/andrewpillar/req/syntax"
)

// File is the value for an open file. This holds the underlying handle to
// the file. The file value can be used as a stream.
type File struct {
	*os.File
}

// ToFile attempts to type assert the given value to a file.
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

// String formats the file to a string. The formatted string will detail the
// pointer at which the underlying handle exists, along with the filename.
func (f File) String() string {
	name := ""

	if f.File != nil {
		name = f.Name()
	}
	return fmt.Sprintf("File<addr=%p, name=%q>", f.File, name)
}

// Sprint returns the entire contents of the underlying file as a string. Once
// read, the file cursor is returned to the beginning of the file.
func (f File) Sprint() string {
	_, _ = f.Seek(0, io.SeekStart)

	b, err := io.ReadAll(f.File)

	if err != nil {
		return ""
	}

	_, _ = f.Seek(0, io.SeekStart)

	return string(b)
}

func (f File) valueType() valueType {
	return fileType
}

func (f File) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, fileType)
}
