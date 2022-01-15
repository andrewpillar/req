package eval

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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

// PrintCmd formats the given values using the Sprint method and writes it to
// standard output. If the final argument given is to a File, then the output
// is written to that file.
var PrintCmd = &Command{
	Name: "print",
	Argc: -1,
	Func: print(os.Stdout),
}

func print(out io.Writer) CommandFunc {
	return func(cmd string, args []value.Value) (value.Value, error) {
		if len(args) < 1 {
			return nil, &CommandError{
				Op:  "call",
				Cmd: cmd,
				Err: errNotEnoughArgs,
			}
		}

		end := len(args) - 1
		last := args[end]

		if f, ok := last.(value.File); ok {
			out = f.File
			args = args[:end]
			end--
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
			case value.String:
				body = strings.NewReader(v.Sprint())
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
		"base64": {
			Argc: 1,
			Func: encodeBase64,
		},
		"form-data": {
			Argc: 1,
			Func: encodeFormData(""),
		},
		"json": {
			Argc: 1,
			Func: encodeJson,
		},
		"url": {
			Argc: 1,
			Func: encodeUrl,
		},
	}
)

func encodeBase64(cmd string, args []value.Value) (value.Value, error) {
	arg0 := args[0]

	var src bytes.Buffer

	switch v := arg0.(type) {
	case value.String:
		src.WriteString(v.Value)
	case value.Stream:
		if _, err := io.Copy(&src, v); err != nil {
			return nil, &CommandError{
				Cmd: cmd,
				Err: err,
			}
		}

		if _, err := v.Seek(0, io.SeekStart); err != nil {
			return nil, &CommandError{
				Cmd: cmd,
				Err: err,
			}
		}
	default:
		return nil, &CommandError{
			Cmd: cmd,
			Err: errors.New("cannot encode " + value.Type(arg0)),
		}
	}

	var buf bytes.Buffer

	enc := base64.NewEncoder(base64.StdEncoding, &buf)

	if _, err := io.Copy(enc, &src); err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	return value.String{
		Value: buf.String(),
	}, nil
}

func encodeFormData(boundary string) CommandFunc {
	return func(cmd string, args []value.Value) (value.Value, error) {
		arg0 := args[0]

		obj, err := value.ToObject(arg0)

		if err != nil {
			return nil, &CommandError{
				Cmd: cmd,
				Err: errors.New("cannot encode " + value.Type(arg0)),
			}
		}

		var buf bytes.Buffer

		w := multipart.NewWriter(&buf)

		if boundary != "" {
			w.SetBoundary(boundary)
		}

		for k, v := range obj.Pairs {
			switch v2 := v.(type) {
			case value.String, value.Int, value.Bool:
				w.WriteField(k, v.Sprint())
			case value.File:
				sw, err := w.CreateFormFile(k, v2.Name())

				if err != nil {
					return nil, &CommandError{
						Cmd: cmd,
						Err: err,
					}
				}

				if _, err := io.Copy(sw, v2); err != nil {
					return nil, &CommandError{
						Cmd: cmd,
						Err: err,
					}
				}

				if _, err := v2.Seek(0, io.SeekStart); err != nil {
					return nil, &CommandError{
						Cmd: cmd,
						Err: err,
					}
				}
			default:
				return nil, &CommandError{
					Cmd: cmd,
					Err: errors.New("key error " + k + ": cannot encode " + value.Type(v)),
				}
			}
		}

		if err := w.Close(); err != nil {
			return nil, &CommandError{
				Cmd: cmd,
				Err: err,
			}
		}

		data := bytes.NewReader(buf.Bytes())

		return &value.FormData{
			Data:        data,
			ContentType: w.FormDataContentType(),
		}, nil
	}
}

func encodeUrl(cmd string, args []value.Value) (value.Value, error) {
	arg0 := args[0]

	obj, err := value.ToObject(arg0)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: errors.New("cannot encode " + value.Type(arg0)),
		}
	}

	vals := make(url.Values)

	for k, v := range obj.Pairs {
		switch v := v.(type) {
		case value.Int, value.Bool, value.String:
			vals[k] = append(vals[k], v.Sprint())
		case *value.Array:
			for _, it := range v.Items {
				vals[k] = append(vals[k], it.Sprint())
			}
		default:
			return nil, &CommandError{
				Cmd: cmd,
				Err: errors.New("key error " + k + ": cannot encode " + value.Type(v)),
			}
		}
	}

	return value.String{
		Value: vals.Encode(),
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

	subcmd.Name = cmd + " " + name.Value

	return subcmd.invoke(args[1:])
}

var (
	DecodeCmd = &Command{
		Name: "decode",
		Argc: 2,
		Func: decode,
	}

	decodetab = map[string]*Command{
		"base64": {
			Argc: 1,
			Func: decodeBase64,
		},
		"form-data": {
			Argc: 1,
			Func: decodeFormData,
		},
		"json": {
			Argc: 1,
			Func: decodeJson,
		},
		"url": {
			Argc: 1,
			Func: decodeUrl,
		},
	}
)

func decodeBase64(cmd string, args []value.Value) (value.Value, error) {
	arg0 := args[0]

	var r io.Reader

	switch v := arg0.(type) {
	case value.String:
		r = strings.NewReader(v.Value)
	case value.Stream:
		r = v
	default:
		return nil, &CommandError{
			Cmd: cmd,
			Err: errors.New("cannot decode " + value.Type(arg0)),
		}
	}

	dec := base64.NewDecoder(base64.StdEncoding, r)

	b, err := io.ReadAll(dec)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}
	return value.NewStream(value.BufferStream(bytes.NewReader(b))), nil
}

var maxFormMemory int64 = 64 << 20 // 64 MB

func decodeFormData(cmd string, args []value.Value) (value.Value, error) {
	arg0 := args[0]

	f, err := value.ToFormData(arg0)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: errors.New("cannot decode " + value.Type(arg0)),
		}
	}

	_, params, err := mime.ParseMediaType(f.ContentType)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	r := multipart.NewReader(f.Data, params["boundary"])

	obj := value.Object{
		Pairs: make(map[string]value.Value),
	}

	form, err := r.ReadForm(maxFormMemory)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	for k, v := range form.Value {
		obj.Pairs[k] = value.String{
			Value: v[0],
		}
	}

	for k, v := range form.File {
		f, err := v[0].Open()

		if err != nil {
			return nil, &CommandError{
				Cmd: cmd,
				Err: err,
			}
		}
		obj.Pairs[k] = value.NewStream(f)
	}
	return obj, nil
}

func decodeJson(cmd string, args []value.Value) (value.Value, error) {
	arg0 := args[0]

	var (
		r      io.Reader
		stream value.Stream
	)

	switch v := arg0.(type) {
	case value.String:
		r = strings.NewReader(v.Value)
	case value.Stream:
		r = v
		stream = v
	default:
		return nil, errors.New("cannot decode " + value.Type(arg0))
	}

	val, err := value.DecodeJSON(r)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	if stream != nil {
		if _, err := stream.Seek(0, io.SeekStart); err != nil {
			return nil, &CommandError{
				Cmd: cmd,
				Err: err,
			}
		}
	}
	return val, nil
}

func decodeUrl(cmd string, args []value.Value) (value.Value, error) {
	arg0 := args[0]

	str, err := value.ToString(arg0)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: errors.New("cannnot decode " + value.Type(arg0)),
		}
	}

	vals, err := url.ParseQuery(str.Value)

	if err != nil {
		return nil, &CommandError{
			Cmd: cmd,
			Err: err,
		}
	}

	obj := value.Object{
		Pairs: make(map[string]value.Value),
	}

	booltab := map[string]bool{
		"true":  true,
		"false": false,
	}

	for k, items := range vals {
		l := len(items)

		vals := make([]value.Value, 0, l)

		for _, it := range items {
			if b, ok := booltab[it]; ok {
				vals = append(vals, value.Bool{Value: b})
				continue
			}

			if '0' >= it[0] && it[0] <= '9' {
				i, err := strconv.ParseInt(it, 10, 64)

				if err != nil {
					vals = append(vals, value.Int{Value: i})
					continue
				}
			}
			vals = append(vals, value.String{Value: it})
		}

		if len(vals) > 1 {
			arr := &value.Array{Items: vals}

			obj.Pairs[k] = arr
			continue
		}
		obj.Pairs[k] = vals[0]
	}
	return obj, nil
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

	subcmd.Name = cmd + " " + name.Value

	return subcmd.invoke(args[1:])
}
