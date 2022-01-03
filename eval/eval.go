package eval

import (
	"bytes"
	"errors"
	"hash/fnv"
	"strconv"

	"github.com/andrewpillar/req/syntax"
	"github.com/andrewpillar/req/token"
)

type symtab struct {
	tab map[string]Object
}

func (s *symtab) put(name nameObj, obj Object) {
	if s.tab == nil {
		s.tab = make(map[string]Object)
	}
	s.tab[name.value] = obj
}

func (s *symtab) get(name nameObj) (Object, error) {
	if s.tab == nil {
		return nil, errors.New("undefined: $" + name.value)
	}

	obj, ok := s.tab[name.value]

	if !ok {
		return nil, errors.New("undefined: $" + name.value)
	}
	return obj, nil
}

type Error struct {
	Pos token.Pos
	Err error
}

func (e Error) Unwrap() error { return e.Err }

func (e Error) Error() string { return e.Pos.String() + " - " + e.Err.Error() }

type Evaluator struct {
	cmds   map[string]*Command
	symtab symtab
}

func (e *Evaluator) AddCmd(cmd *Command) {
	if e.cmds == nil {
		e.cmds = make(map[string]*Command)
	}
	e.cmds[cmd.Name] = cmd
}

func (e *Evaluator) interpolate(pos token.Pos, s string) (Object, error) {
	var buf bytes.Buffer

	interpolate := false
	expr := make([]rune, 0, len(s))

	for _, r := range s {
		if r == '{' {
			interpolate = true
			continue
		}

		if r == '}' {
			interpolate = false

			n, err := syntax.ParseRef(string(expr))

			if err != nil {
				err = errors.Unwrap(err)

				return nil, err
			}

			obj, err := e.Eval(n)

			if err != nil {
				return nil, err
			}

			buf.WriteString(obj.String())
			expr = expr[0:0]
			continue
		}

		if interpolate {
			expr = append(expr, r)
			continue
		}
		buf.WriteRune(r)
	}
	return stringObj{value: buf.String()}, nil
}

func (e *Evaluator) resolveCommand(n *syntax.CommandStmt) (*Command, []Object, error) {
	cmd, ok := e.cmds[n.Name.Value]

	if !ok {
		return nil, nil, errors.New("undefined command: " + n.Name.Value)
	}

	args := make([]Object, 0, len(n.Args))

	for _, arg := range n.Args {
		obj, err := e.Eval(arg)

		if err != nil {
			return nil, nil, e.err(arg.Pos(), err)
		}
		args = append(args, obj)
	}
	return cmd, args, nil
}

func (e *Evaluator) resolveArrayIndex(arr, ind Object) (Object, error) {
	i64, ok := ind.(intObj)

	if !ok {
		return nil, TypeError{
			typ:      ind.Type(),
			expected: Int,
		}
	}

	arrobj := arr.(arrayObj)
	end := len(arrobj.items) - 1

	i := int(i64.value)

	if i < 0 || i > end {
		return nil, nil
	}
	return arrobj.items[i], nil
}

func (e *Evaluator) resolveHashKey(hash, key Object) (Object, error) {
	s, ok := key.(stringObj)

	if !ok {
		return nil, TypeError{
			typ:      key.Type(),
			expected: String,
		}
	}

	obj, ok := hash.(hashObj).pairs[s.value]

	if !ok {
		return nil, nil
	}
	return obj, nil
}

func (e *Evaluator) resolveDot(n *syntax.DotExpr) (Object, error) {
	left, err := e.Eval(n.Left)

	if err != nil {
		return nil, err
	}

	name, ok := left.(nameObj)

	if !ok {
		return nil, errors.New("cannot use type " + left.Type().String() + " as selector")
	}

	obj, err := e.symtab.get(name)

	if err != nil {
		return nil, err
	}

	sel, ok := obj.(Selector)

	if !ok {
		return nil, errors.New("cannot use type " + obj.Type().String() + " as selector")
	}

	right, err := e.Eval(n.Right)

	if err != nil {
		return nil, err
	}

	obj, err = sel.Select(right)

	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (e *Evaluator) resolveInd(n *syntax.IndExpr) (Object, error) {
	left, err := e.Eval(n.Left)

	if err != nil {
		return nil, err
	}

	right, err := e.Eval(n.Right)

	if err != nil {
		return nil, err
	}

	switch left.Type() {
	case Array:
		return e.resolveArrayIndex(left, right)
	case Hash:
		return e.resolveHashKey(left, right)
	default:
		return nil, errors.New("type " + left.Type().String() + " does not support indexing")
	}
}

func (e *Evaluator) err(pos token.Pos, err error) error {
	if _, ok := err.(Error); ok {
		return err
	}

	return Error{
		Pos: pos,
		Err: err,
	}
}

func (e *Evaluator) Eval(n syntax.Node) (Object, error) {
	switch v := n.(type) {
	case *syntax.VarDecl:
		name := nameObj{value: v.Name.Value}

		obj, err := e.Eval(v.Value)

		if err != nil {
			return nil, err
		}
		e.symtab.put(name, obj)
	case *syntax.Ref:
		switch v := v.Left.(type) {
		case *syntax.Name:
			name := nameObj{value: v.Value}
			return e.symtab.get(name)
		case *syntax.DotExpr:
			return e.resolveDot(v)
		case *syntax.IndExpr:
			return e.resolveInd(v)
		default:
			return nil, errors.New("invalid reference")
		}
	case *syntax.DotExpr:
		return e.resolveDot(v)
	case *syntax.IndExpr:
		return e.resolveInd(v)
	case *syntax.Lit:
		switch v.Type {
		case token.String:
			return e.interpolate(v.Pos(), v.Value)
		case token.Int:
			i, _ := strconv.ParseInt(v.Value, 10, 64)
			return intObj{value: i}, nil
		case token.Bool:
			b := true

			if v.Value != "true" {
				b = false
			}
			return boolObj{value: b}, nil
		}
	case *syntax.Name:
		return nameObj{value: v.Value}, nil
	case *syntax.Array:
		items := make([]Object, 0, len(v.Items))

		for _, it := range v.Items {
			obj, err := e.Eval(it)

			if err != nil {
				return nil, err
			}
			items = append(items, obj)
		}
		return arrayObj{items: items}, nil
	case *syntax.Object:
		pairs := make(map[string]Object)

		for _, n := range v.Pairs {
			obj, err := e.Eval(n.Value)

			if err != nil {
				return nil, err
			}
			pairs[n.Key.Value] = obj
		}
		return hashObj{pairs: pairs}, nil
	case *syntax.BlockStmt:
		for _, n := range v.Nodes {
			if _, err := e.Eval(n); err != nil {
				return nil, err
			}
		}
		return nil, nil
	case *syntax.CommandStmt:
		cmd, args, err := e.resolveCommand(v)

		if err != nil {
			return nil, e.err(n.Pos(), err)
		}

		obj, err := cmd.Invoke(args)

		if err != nil {
			return nil, err
		}
		return obj, nil
	case *syntax.MatchStmt:
		obj, err := e.Eval(v.Cond)

		if err != nil {
			return nil, err
		}

		jmptab := make(map[uint32]syntax.Node)

		for _, stmt := range v.Cases {
			h := fnv.New32a()

			obj, err := e.Eval(stmt.Value)

			if err != nil {
				return nil, err
			}

			h.Write([]byte(obj.String()))

			jmptab[h.Sum32()] = stmt.Then
		}

		if typ := obj.Type(); typ != String && typ != Int {
			return nil, errors.New("cannot match against type " + typ.String())
		}

		h := fnv.New32a()
		h.Write([]byte(obj.String()))

		if n, ok := jmptab[h.Sum32()]; ok {
			return e.Eval(n)
		}

		if v.Default != nil {
			return e.Eval(v.Default)
		}
		return nil, nil
	case *syntax.ChainExpr:
		var obj Object

		for _, n := range v.Commands {
			cmd, args, err := e.resolveCommand(n)

			if err != nil {
				return nil, err
			}

			if obj != nil {
				args = append([]Object{obj}, args...)
			}

			obj, err = cmd.Invoke(args)

			if err != nil {
				return nil, err
			}
		}
		return obj, nil
	case *syntax.IfStmt:
	}
	return nil, nil
}

func (e *Evaluator) Run(nn []syntax.Node) error {
	for _, n := range nn {
		if _, err := e.Eval(n); err != nil {
			return err
		}
	}
	return nil
}
