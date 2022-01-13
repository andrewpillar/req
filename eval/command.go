package eval

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/req/value"
	"github.com/andrewpillar/req/version"
)

// CommandFunc is the function for handling the invocation of a command. This
// is passed the name of the command being invoked, and the list of arguments
// given.
type CommandFunc func(cmd string, args []value.Value) (value.Value, error)

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

func (e *CommandError) Error() string {
	if e.Op != "" {
		return "invalid " + e.Op + " to " + e.Cmd + ": " + e.Err.Error()
	}
	return e.Cmd + ": " + e.Err.Error()
}

// invoke executes the command. Before execution it will ensure the number of
// arguments given the amount the command expects, otherwise this will return
// a CommandError.
func (c Command) invoke(args []value.Value) (value.Value, error) {
	if c.Argc > -1 {
		if l := len(args); l != c.Argc {
			if l > c.Argc {
				return nil, &CommandError{
					Op:  "call",
					Cmd: c.Name,
					Err: errTooManyArgs,
				}
			}

			return nil, &CommandError{
				Op:  "call",
				Cmd: c.Name,
				Err: errNotEnoughArgs,
			}
		}
	}
	return c.Func(c.Name, args)
}

// EnvCmd is for the "env" command that allows for retrieving environment
// variables. This takes a single argument that is the name of the variable.
// This returns a string for the environment variable.
var EnvCmd = &Command{
	Name: "env",
	Argc: 1,
	Func: env,
}

func env(cmd string, args []value.Value) (value.Value, error) {
	str, err := value.ToString(args[0])

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	return value.String{
		Value: os.Getenv(str.Value),
	}, nil
}

// ExitCmd is for the "exit" command that allows for exiting a script. This
// takes a single argument which is the exit code to use.
var ExitCmd = &Command{
	Name: "exit",
	Argc: 1,
	Func: exit,
}

func exit(cmd string, args []value.Value) (value.Value, error) {
	i, err := value.ToInt(args[0])

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	os.Exit(int(i.Value))
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

func open(cmd string, args []value.Value) (value.Value, error) {
	str, err := value.ToString(args[0])

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	if err := os.MkdirAll(filepath.Dir(str.Value), os.FileMode(0755)); err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	f, err := os.OpenFile(str.Value, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.FileMode(0644))

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	return value.File{
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

func print_(cmd string, args []value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, &CommandError{
			Op:  "call",
			Cmd: cmd,
			Err: errNotEnoughArgs,
		}
	}

	out := os.Stdout

	end := len(args) - 1
	last := args[end]

	if f, ok := last.(value.File); ok {
		out = f.File
		args = args[:end]
	}

	var buf bytes.Buffer

	for i, arg := range args {
		if _, err := fmt.Fprint(&buf, arg.Sprint()); err != nil {
			return nil, &CommandError{
				Cmd: cmd,
				Err: err,
			}
		}

		if i != end {
			buf.WriteByte(' ')
		}
	}

	buf.WriteByte('\n')

	if _, err := fmt.Fprint(out, buf.String()); err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}
	return nil, nil
}

var (
	HeadCmd = &Command{
		Name: "HEAD",
		Argc: -1,
		Func: func(cmd string, args []value.Value) (value.Value, error) {
			if len(args) < 1 {
				return nil, &CommandError{
					Op:  "call",
					Cmd: cmd,
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
		Func: func(cmd string, args []value.Value) (value.Value, error) {
			if len(args) < 1 {
				return nil, &CommandError{
					Op:  "call",
					Cmd: cmd,
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
		Func: func(cmd string, args []value.Value) (value.Value, error) {
			if len(args) < 1 {
				return nil, &CommandError{
					Op:  "call",
					Cmd: cmd,
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
		Func: func(cmd string, args []value.Value) (value.Value, error) {
			if len(args) < 1 {
				return nil, &CommandError{
					Op:  "call",
					Cmd: cmd,
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
		Func: func(cmd string, args []value.Value) (value.Value, error) {
			if len(args) < 1 {
				return nil, &CommandError{
					Op:  "call",
					Cmd: cmd,
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
		Func: func(cmd string, args []value.Value) (value.Value, error) {
			if len(args) < 1 {
				return nil, &CommandError{
					Op:  "call",
					Cmd: cmd,
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
		Func: func(cmd string, args []value.Value) (value.Value, error) {
			if len(args) < 1 {
				return nil, &CommandError{
					Op:  "call",
					Cmd: cmd,
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

func request(cmd string, args []value.Value) (value.Value, error) {
	var body io.Reader

	endpoint, err := value.ToString(args[0])

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	var obj value.Object

	if len(args) > 1 {
		obj, err = value.ToObject(args[1])

		if err != nil {
			return nil, &CommandError{
				Cmd: cmd,
				Err: err,
			}
		}

		if len(args) > 2 {
			arg2 := args[2]

			switch v := arg2.(type) {
			case value.String, *value.Array, value.Object:
				body = strings.NewReader(arg2.Sprint())
			case value.File:
				body = v.File
			default:
				return nil, &CommandError{
					Cmd: cmd,
					Err: errors.New("cannot use type " + value.Type(arg2) + " as request body"),
				}
			}
		}
	}

	req, err := http.NewRequest(cmd, endpoint.Value, body)

	if err != nil {
		return nil, err
	}

	for key, val := range obj.Pairs {
		str, err := value.ToString(val)

		if err != nil {
			return nil, &CommandError{
				Cmd: cmd,
				Err: err,
			}
		}
		req.Header.Set(key, str.Value)
	}

	if val := req.Header.Get("User-Agent"); val == "" {
		req.Header.Set("User-Agent", "req/"+version.Build)
	}

	return value.Request{
		Request: req,
	}, nil
}

var SendCmd = &Command{
	Name: "send",
	Argc: 1,
	Func: send,
}

func send(cmd string, args []value.Value) (value.Value, error) {
	req, err := value.ToRequest(args[0])

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	var cli http.Client

	resp, err := cli.Do(req.Request)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	return value.Response{
		Response: resp,
	}, nil
}

// SniffCmd is for the "sniff" command that allows for inspecting the mime type
// of a stream. This takes a single argument, and returns a string.
var SniffCmd = &Command{
	Name: "sniff",
	Argc: 1,
	Func: sniff,
}

func sniff(cmd string, args []value.Value) (value.Value, error) {
	s, err := value.ToStream(args[0])

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	hdr := make([]byte, 512)

	if _, err := s.Read(hdr); err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	if _, err := s.Seek(0, io.SeekStart); err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	return value.String{
		Value: http.DetectContentType(hdr),
	}, nil
}

var (
	EncodeCmd = &Command{
		Name: "encode",
		Argc: 2,
		Func: encode,
	}

	encodetab = map[string]*Command{
		"base64": &Command{
			Argc: 1,
			Func: encodeBase64,
		},
		"json": &Command{
			Argc: 1,
			Func: encodeJson,
		},
	}

	DecodeCmd = &Command{
		Name: "decode",
		Argc: 2,
		Func: decode,
	}

	decodetab = map[string]*Command{
		"base64": &Command{
			Argc: 1,
			Func: decodeBase64,
		},
		"json": &Command{
			Argc: 1,
			Func: decodeJson,
		},
	}
)

func encodeBase64(cmd string, args []value.Value) (value.Value, error) {
	arg0 := args[0]

	var src []byte

	switch v := arg0.(type) {
	case value.String:
		src = []byte(v.Value)
	case value.File:
		b, err := io.ReadAll(v.File)

		if err != nil {
			return nil, &CommandError{
				Cmd: cmd,
				Err: err,
			}
		}

		if _, err := v.File.Seek(0, io.SeekStart); err != nil {
			return nil, &CommandError{
				Cmd: cmd,
				Err: err,
			}
		}
		src = b
	default:
		return nil, &CommandError{
			Cmd: cmd,
			Err: errors.New("cannot encode " + value.Type(arg0)),
		}
	}

	return value.String{
		Value: base64.StdEncoding.EncodeToString(src),
	}, nil
}

func encodeJson(cmd string, args []value.Value) (value.Value, error) {
	arg0 := args[0]

	switch arg0.(type) {
	case *value.Array:
	case value.Object:
	default:
		return nil, &CommandError{
			Cmd: cmd,
			Err: errors.New("cannot encode " + value.Type(arg0)),
		}
	}

	b, err := json.Marshal(arg0)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	return value.String{
		Value: string(b),
	}, nil
}

func encode(cmd string, args []value.Value) (value.Value, error) {
	name, err := value.ToName(args[0])

	if err != nil {
		return nil, err
	}

	subcmd, ok := encodetab[name.Value]

	if !ok {
		return nil, errors.New("undefined command: " + cmd + " " + name.Value)
	}

	subcmd.Name = cmd  + " " + name.Value

	return subcmd.invoke(args[1:])
}

func decodeBase64(cmd string, args []value.Value) (value.Value, error) {
	arg0 := args[0]

	str, err := value.ToString(arg0)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: errors.New("cannot decode " + value.Type(arg0)),
		}
	}

	b, err := base64.StdEncoding.DecodeString(str.Value)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	return value.String{
		Value: string(b),
	}, nil
}

func decodeJson(cmd string, args []value.Value) (value.Value, error) {
	arg0 := args[0]

	var (
		r      io.Reader
		rewind func() error
	)

	switch v := arg0.(type) {
	case value.String:
		r = strings.NewReader(v.Value)
	case value.File:
		r = v.File
		rewind = func() error {
			if _, err := v.File.Seek(0, io.SeekStart); err != nil {
				return err
			}
			return nil
		}
	case value.Stream:
		r = v
		rewind = func() error {
			if _, err := v.Seek(0, io.SeekStart); err != nil {
				return err
			}
			return nil
		}
	default:
		return nil, errors.New("cannot decode " + value.Type(arg0))
	}

	if rewind != nil {
		if err := rewind(); err != nil {
			return nil, &CommandError{
				Cmd: cmd,
				Err: err,
			}
		}
	}

	val, err := value.DecodeJSON(r)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}
	return val, nil
}

func decode(cmd string, args []value.Value) (value.Value, error) {
	name, err := value.ToName(args[0])

	if err != nil {
		return nil, err
	}

	subcmd, ok := decodetab[name.Value]

	if !ok {
		return nil, errors.New("undefined command: " + cmd + " " + name.Value)
	}

	subcmd.Name = cmd  + " " + name.Value

	return subcmd.invoke(args[1:])
}
