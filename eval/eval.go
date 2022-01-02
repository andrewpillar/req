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
		return nil, errors.New("undefined: " + name.value)
	}

	obj, ok := s.tab[name.value]

	if !ok {
		return nil, errors.New("undefined: " + name.value)
	}
	return obj, nil
}

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
		}

		if r == '}' {
			interpolate = false
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
		return nil, nil, n.Err("undefined command: " + n.Name.Value)
	}

	args := make([]Object, 0, len(n.Args))

	for _, arg := range n.Args {
		obj, err := e.Eval(arg)

		if err != nil {
			return nil, nil, arg.Err(err.Error())
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
		return nil, n.Left.Err(err.Error())
	}

	name, ok := left.(nameObj)

	if !ok {
		return nil, n.Left.Err("expected name")
	}

	obj, err := e.symtab.get(name)

	if err != nil {
		return nil, n.Err(err.Error())
	}

	sel, ok := obj.(Selector)

	if !ok {
		return nil, errors.New("cannot use type " + obj.Type().String() + " as selector")
	}

	right, err := e.Eval(n.Right)

	if err != nil {
		return nil, n.Right.Err(err.Error())
	}

	obj, err = sel.Select(right)

	if err != nil {
		return nil, n.Err(err.Error())
	}
	return obj, nil
}

func (e *Evaluator) Eval(n syntax.Node) (Object, error) {
	switch v := n.(type) {
	case *syntax.VarDecl:
		name := nameObj{value: v.Name.Value}

		obj, err := e.Eval(v.Value)

		if err != nil {
			return nil, v.Err(err.Error())
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
			left, err := e.Eval(v.Left)

			if err != nil {
				return nil, v.Err(err.Error())
			}

			right, err := e.Eval(v.Right)

			if err != nil {
				return nil, v.Err(err.Error())
			}

			switch left.Type() {
			case Array:
				obj, err := e.resolveArrayIndex(left, right)

				if err != nil {
					return nil, v.Err(err.Error())
				}
				return obj, err
			case Hash:
				obj, err := e.resolveHashKey(left, right)

				if err != nil {
					return nil, v.Err(err.Error())
				}
				return obj, err
			default:
				return nil, v.Left.Err("type " + left.Type().String() + " does not support indexing")
			}
		default:
			return nil, v.Err("invalid reference")
		}
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
				return nil, it.Err(err.Error())
			}
			items = append(items, obj)
		}
		return arrayObj{items: items}, nil
	case *syntax.Object:
		pairs := make(map[string]Object)

		for _, n := range v.Pairs {
			obj, err := e.Eval(n.Value)

			if err != nil {
				return nil, n.Value.Err(err.Error())
			}
			pairs[n.Key.Value] = obj
		}
		return hashObj{pairs: pairs}, nil
	case *syntax.BlockStmt:
		for _, n := range v.Nodes {
			if _, err := e.Eval(n); err != nil {
				return nil, n.Err(err.Error())
			}
		}
		return nil, nil
	case *syntax.CommandStmt:
		cmd, args, err := e.resolveCommand(v)

		if err != nil {
			return nil, err
		}

		obj, err := cmd.Invoke(args)

		if err != nil {
			return nil, v.Err(err.Error())
		}
		return obj, nil
	case *syntax.MatchStmt:
		obj, err := e.Eval(v.Cond)

		if err != nil {
			return nil, v.Err(err.Error())
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
			return nil, v.Err("cannot match against type " + typ.String())
		}

		h := fnv.New32a()
		h.Write([]byte(obj.String()))

		if n, ok := jmptab[h.Sum32()]; ok {
			return e.Eval(n)
		}
		return nil, nil
	case *syntax.YieldStmt:
		obj, err := e.Eval(v.Value)

		if err != nil {
			return nil, v.Err(err.Error())
		}
		return obj, nil
	case *syntax.ChainExpr:
		var obj Object

		for _, n := range v.Commands {
			cmd, args, err := e.resolveCommand(n)

			if err != nil {
				return nil, n.Err(err.Error())
			}

			if obj != nil {
				args = append([]Object{obj}, args...)
			}

			obj, err = cmd.Invoke(args)

			if err != nil {
				return nil, n.Err(err.Error())
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
