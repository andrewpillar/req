package eval

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type CommandFunc func(args []Object) (Object, error)

type Command struct {
	Name string
	Argc int
	Args []Object
	Func CommandFunc
}

type CommandError struct {
	Op  string
	Cmd string
	Err error
}

var errNotEnoughArgs = errors.New("not enough arguments")

func (e CommandError) Error() string {
	if e.Op != "" {
		return "invalid " + e.Op + " to " + e.Cmd + ": " + e.Err.Error() 
	}
	return e.Cmd + ": " + e.Err.Error()
}

func (c Command) Invoke(args []Object) (Object, error) {
	if c.Argc > -1 {
		if l := len(args); l != c.Argc {
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

	f, err := os.OpenFile(str.value, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0644))

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

var PrintCmd = &Command{
	Name: "write",
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

	last := args[len(args)-1]

	if f, ok := last.(fileObj); ok {
		out = f.File
	}

	var buf bytes.Buffer

	for _, arg := range args {
		buf.WriteString(arg.String())
	}

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

		if len(args) >= 2 {
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
