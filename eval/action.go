package eval

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

type ActionFunc func(args []Value, dest Value) (Value, error)

type Action struct {
	Name string
	Argc int
	Func ActionFunc
}

func (a Action) Call(args []Value, dest Value) (Value, error) {
	if a.Argc > -1 {
		if l := len(args); l != a.Argc {
			if l < a.Argc {
				return nil, errors.New("too few arguments in call to " + a.Name)
			}
			return nil, errors.New("not enough arguments in call to " + a.Name)
		}
	}
	return a.Func(args, dest)
}

var EnvAction = Action{
	Name: "env",
	Argc: 1,
	Func: Env,
}

func Env(args []Value, _ Value) (Value, error) {
	val := args[0]

	str, ok := val.(String)

	if !ok {
		return nil, errors.New("cannot use type " + val.type_().String() + " as type string")
	}

	return String{
		Value: os.Getenv(str.Value),
	}, nil
}

var ExitAction = Action{
	Name: "exit",
	Argc: 1,
	Func: Exit,
}

func Exit(args []Value, _ Value) (Value, error) {
	val := args[0]

	i, ok := val.(Int)

	if !ok {
		return nil, errors.New("cannot use type " + val.type_().String() + " as type int")
	}

	os.Exit(int(i.Value))
	return nil, nil
}

var OpenAction = Action{
	Name: "open",
	Argc: 1,
	Func: Open,
}

func Open(args []Value, _ Value) (Value, error) {
	val := args[0]

	str, ok := val.(String)

	if !ok {
		return nil, errors.New("cannot use type " + val.type_().String() + " as type string")
	}

	f, err := os.Open(str.Value)

	if err != nil {
		return nil, err
	}
	return File{
		File: f,
	}, nil
}

var WriteAction = Action{
	Name: "write",
	Argc: -1,
	Func: Write,
}

func Write(args []Value, dest Value) (Value, error) {
	f, ok := dest.(File)

	if !ok {
		return nil, errors.New("cannot use type " + dest.type_().String() + " as type file")
	}

	for _, arg := range args {
		if _, err := f.WriteString(arg.String()); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func method(name string, args []Value, dest Value) (Value, error) {
	str, ok := dest.(String)

	if !ok {
		return nil, errors.New("cannot use type " + dest.type_().String() + " as type string")
	}

	var (
		cli  http.Client
		body io.Reader
	)

	if len(args) > 1 {
		switch v := args[1].(type) {
		case String, Array, Object:
			body = strings.NewReader(v.String())
		case File:
			body = v.File
		default:
			return nil, errors.New("cannot use type " + v.type_().String() + " as request body")
		}
	}

	req, err := http.NewRequest(name, str.Value, body)

	if err != nil {
		return nil, err
	}

	defer req.Body.Close()

	if len(args) > 0 {
		arg := args[0]

		obj, ok := arg.(Object)

		if !ok {
			return nil, errors.New("cannot use type " + arg.type_().String() + " as type object")
		}

		for key, val := range obj.Pairs {
			str, ok := val.(String)

			if !ok {
				return nil, errors.New("cannot use type " + val.type_().String() + " as type string")
			}
			req.Header.Set(key.Name, str.Value)
		}
	}

	resp, err := cli.Do(req)

	if err != nil {
		return nil, err
	}

	return Resp{
		Response: resp,
	}, nil
}

var HeadAction = Action{
	Name: "HEAD",
	Argc: -1,
	Func: Head,
}

func Head(args []Value, dest Value) (Value, error) {
	if len(args) > 1 {
		args = args[:1]
	}
	return method("HEAD", args, dest)
}

var OptionsAction = Action{
	Name: "OPTIONS",
	Argc: -1,
	Func: Options,
}

func Options(args []Value, dest Value) (Value, error) {
	if len(args) > 1 {
		args = args[:1]
	}
	return method("OPTIONS", args, dest)
}

var GetAction = Action{
	Name: "GET",
	Argc: -1,
	Func: Get,
}

func Get(args []Value, dest Value) (Value, error) {
	if len(args) > 1 {
		args = args[:1]
	}
	return method("GET", args, dest)
}

var PostAction = Action{
	Name: "POST",
	Argc: -1,
	Func: Post,
}

func Post(args []Value, dest Value) (Value, error) {
	if len(args) > 2 {
		args = args[:2]
	}
	return method("POST", args, dest)
}

var PutAction = Action{
	Name: "PUT",
	Argc: -1,
	Func: Put,
}

func Put(args []Value, dest Value) (Value, error) {
	if len(args) > 2 {
		args = args[:2]
	}
	return method("PUT", args, dest)
}

var PatchAction = Action{
	Name: "PATCH",
	Argc: -1,
	Func: Patch,
}

func Patch(args []Value, dest Value) (Value, error) {
	if len(args) > 2 {
		args = args[:2]
	}
	return method("PATCH", args, dest)
}

var DeleteAction = Action{
	Name: "DELETE",
	Argc: -1,
	Func: Delete,
}

func Delete(args []Value, dest Value) (Value, error) {
	if len(args) > 1 {
		args = args[:1]
	}
	return method("DELETE", args, dest)
}
