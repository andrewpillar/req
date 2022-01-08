package eval

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// CommandFunc is the function for handling the invocation of a command.
type CommandFunc func(args []Object) (Object, error)

type Command struct {
	Name string      // The name of the command.
	Argc int         // The number of arguments for the command, -1 for unlimited.
	Func CommandFunc // The function to execute for the command.
}

// CommandError records an error and the operation and command that caused it.
type CommandError struct {
	Op  string
	Cmd string
	Err error
}

var (
	errNotEnoughArgs = errors.New("not enough arguments")
	errTooManyArgs   = errors.New("too many arguments")
)

func (e CommandError) Error() string {
	if e.Op != "" {
		return "invalid " + e.Op + " to " + e.Cmd + ": " + e.Err.Error()
	}
	return e.Cmd + ": " + e.Err.Error()
}

// Invoke executes the command. Before execution it will ensure the number of
// arguments given the amount the command expects, otherwise this will return
// a CommandError.
func (c Command) Invoke(args []Object) (Object, error) {
	if c.Argc > -1 {
		if l := len(args); l != c.Argc {
			if l > c.Argc {
				return nil, CommandError{
					Op:  "call",
					Cmd: c.Name,
					Err: errTooManyArgs,
				}
			}

			return nil, CommandError{
				Op:  "call",
				Cmd: c.Name,
				Err: errNotEnoughArgs,
			}
		}
	}
	return c.Func(args)
}

type TypeError struct {
	typ      Type
	expected Type
}

func (e TypeError) Error() string {
	return "cannot use " + e.typ.String() + " as type " + e.expected.String()
}

var trueCmd = &Command{
	Name: "true",
	Func: func(_ []Object) (Object, error) {
		return boolObj{value: true}, nil
	},
}

var falseCmd = &Command{
	Name: "false",
	Func: func(_ []Object) (Object, error) {
		return boolObj{}, nil
	},
}

// EnvCmd is for the "env" command that allows for retrieving environment
// variables. This takes a single argument that is the name of the variable.
// This returns a string for the environment variable.
var EnvCmd = &Command{
	Name: "env",
	Argc: 1,
	Func: env,
}

func env(args []Object) (Object, error) {
	val := args[0]

	str, ok := val.(stringObj)

	if !ok {
		return nil, TypeError{
			typ:      val.Type(),
			expected: String,
		}
	}

	return stringObj{
		value: os.Getenv(str.value),
	}, nil
}

// ExitCmd is for the "exit" command that allows for exiting a script. This
// takes a single argument which is the exit code to use.
var ExitCmd = &Command{
	Name: "exit",
	Argc: 1,
	Func: exit,
}

func exit(args []Object) (Object, error) {
	val := args[0]

	i, ok := val.(intObj)

	if !ok {
		return nil, TypeError{
			typ:      val.Type(),
			expected: Int,
		}
	}

	os.Exit(int(i.value))
	return nil, nil
}

// OpenCmd is for the "open" command that allows for opening a file. This takes
// a single argument which is the path of the file to open. This returns a
// handle to that file.
var OpenCmd = &Command{
	Name: "open",
	Argc: 1,
	Func: open,
}

func open(args []Object) (Object, error) {
	val := args[0]

	str, ok := val.(stringObj)

	if !ok {
		return nil, TypeError{
			typ:      val.Type(),
			expected: String,
		}
	}

	if err := os.MkdirAll(filepath.Dir(str.value), os.FileMode(0755)); err != nil {
		return nil, CommandError{
			Cmd: "open",
			Err: err,
		}
	}

	f, err := os.OpenFile(str.value, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.FileMode(0644))

	if err != nil {
		return nil, CommandError{
			Cmd: "open",
			Err: err,
		}
	}

	return fileObj{
		File: f,
	}, nil
}

// PrintCmd is for the "print" command that allows for printing to a file or
// to stdout. This takes an unlimited number of arguments. If the final argument
// is a file, then the output is written to that file. This returns nothing.
var PrintCmd = &Command{
	Name: "print",
	Argc: -1,
	Func: print_,
}

func print_(args []Object) (Object, error) {
	if len(args) < 1 {
		return nil, CommandError{
			Op:  "call",
			Cmd: "print",
			Err: errNotEnoughArgs,
		}
	}

	out := os.Stdout

	end := len(args) - 1
	last := args[end]

	if f, ok := last.(fileObj); ok {
		out = f.File
		args = args[:end]
	}

	var buf bytes.Buffer

	for i, arg := range args {
		buf.WriteString(arg.String())

		if i != end {
			buf.WriteByte(' ')
		}
	}

	buf.WriteByte('\n')

	if _, err := fmt.Fprint(out, buf.String()); err != nil {
		return nil, CommandError{
			Cmd: "print",
			Err: err,
		}
	}
	return nil, nil
}

var (
	HeadCmd = &Command{
		Name: "HEAD",
		Argc: -1,
		Func: func(args []Object) (Object, error) {
			if len(args) < 1 {
				return nil, CommandError{
					Op:  "call",
					Cmd: "HEAD",
					Err: errNotEnoughArgs,
				}
			}

			if len(args) > 2 {
				args = args[:2]
			}
			return request("HEAD", args)
		},
	}

	OptionsCmd = &Command{
		Name: "OPTIONS",
		Argc: -1,
		Func: func(args []Object) (Object, error) {
			if len(args) < 1 {
				return nil, CommandError{
					Op:  "call",
					Cmd: "OPTIONS",
					Err: errNotEnoughArgs,
				}
			}

			if len(args) > 2 {
				args = args[:2]
			}
			return request("OPTIONS", args)
		},
	}

	GetCmd = &Command{
		Name: "GET",
		Argc: -1,
		Func: func(args []Object) (Object, error) {
			if len(args) < 1 {
				return nil, CommandError{
					Op:  "call",
					Cmd: "GET",
					Err: errNotEnoughArgs,
				}
			}

			if len(args) > 2 {
				args = args[:2]
			}
			return request("GET", args)
		},
	}

	PostCmd = &Command{
		Name: "POST",
		Argc: -1,
		Func: func(args []Object) (Object, error) {
			if len(args) < 1 {
				return nil, CommandError{
					Op:  "call",
					Cmd: "POST",
					Err: errNotEnoughArgs,
				}
			}

			if len(args) > 3 {
				args = args[:3]
			}
			return request("POST", args)
		},
	}

	PatchCmd = &Command{
		Name: "PATCH",
		Argc: -1,
		Func: func(args []Object) (Object, error) {
			if len(args) < 1 {
				return nil, CommandError{
					Op:  "call",
					Cmd: "PATCH",
					Err: errNotEnoughArgs,
				}
			}

			if len(args) > 3 {
				args = args[:3]
			}
			return request("PATCH", args)
		},
	}

	PutCmd = &Command{
		Name: "PUT",
		Argc: -1,
		Func: func(args []Object) (Object, error) {
			if len(args) < 1 {
				return nil, CommandError{
					Op:  "call",
					Cmd: "PUT",
					Err: errNotEnoughArgs,
				}
			}

			if len(args) > 3 {
				args = args[:3]
			}
			return request("PUT", args)
		},
	}

	DeleteCmd = &Command{
		Name: "DELETE",
		Argc: -1,
		Func: func(args []Object) (Object, error) {
			if len(args) < 1 {
				return nil, CommandError{
					Op:  "call",
					Cmd: "DELETE",
					Err: errNotEnoughArgs,
				}
			}

			if len(args) > 2 {
				args = args[:2]
			}
			return request("DELETE", args)
		},
	}
)

func request(name string, args []Object) (Object, error) {
	var body io.Reader

	arg0 := args[0]

	endpoint, ok := arg0.(stringObj)

	if !ok {
		return nil, TypeError{
			typ:      arg0.Type(),
			expected: String,
		}
	}

	var hash hashObj

	if len(args) > 1 {
		arg1 := args[1]

		if arg1.Type() != Hash {
			return nil, TypeError{
				typ:      arg1.Type(),
				expected: Hash,
			}
		}

		hash = arg1.(hashObj)

		if len(args) > 2 {
			arg2 := args[2]

			switch arg2.Type() {
			case String, Array, Hash:
				body = strings.NewReader(arg2.String())
			case File:
				f := arg2.(fileObj)

				body = f.File
			default:
				return nil, errors.New("cannot use type " + arg2.Type().String() + " as request body")
			}
		}
	}

	req, err := http.NewRequest(name, endpoint.value, body)

	if err != nil {
		return nil, err
	}

	for key, val := range hash.pairs {
		str, ok := val.(stringObj)

		if !ok {
			return nil, TypeError{
				typ:      val.Type(),
				expected: String,
			}
		}
		req.Header.Set(key, str.value)
	}

	if val := req.Header.Get("User-Agent"); val == "" {
		req.Header.Set("User-Agent", "req")
	}
	return reqObj{
		Request: req,
	}, nil
}

var SendCmd = &Command{
	Name: "send",
	Argc: 1,
	Func: send,
}

func send(args []Object) (Object, error) {
	val := args[0]

	req, ok := val.(reqObj)

	if !ok {
		return nil, TypeError{
			typ:      val.Type(),
			expected: Request,
		}
	}

	var cli http.Client

	resp, err := cli.Do(req.Request)

	if err != nil {
		return nil, CommandError{
			Cmd: "send",
			Err: err,
		}
	}
	return respObj{
		Response: resp,
	}, nil
}

// SniffCmd is for the "sniff" command that allows for inspecting the mime type
// of a file or a stream. This takes a single argument, and returns a string.
var SniffCmd = &Command{
	Name: "sniff",
	Argc: 1,
	Func: sniff,
}

func sniff(args []Object) (Object, error) {
	val := args[0]

	var rs io.ReadSeeker

	switch v := val.(type) {
	case streamObj:
		rs = v.rs
	case fileObj:
		rs = v.File
	default:
		return nil, errors.New("cannot use type " + val.Type().String() + " as stream or file")
	}

	hdr := make([]byte, 512)

	rs.Read(hdr)
	rs.Seek(0, io.SeekStart)

	mime := http.DetectContentType(hdr)

	return stringObj{value: mime}, nil
}
