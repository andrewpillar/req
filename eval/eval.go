package eval

import (
	"errors"
	"strconv"

	"github.com/andrewpillar/req/syntax"
	"github.com/andrewpillar/req/token"
)

type symtab struct {
	tab map[string]Value
}

func (s *symtab) put(name string, val Value) {
	if s.tab == nil {
		s.tab = make(map[string]Value)
	}
	s.tab[name] = val
}

func (s *symtab) get(name string) (Value, error) {
	if s.tab == nil {
		return nil, errors.New("undefined: " + name)
	}

	val, ok := s.tab[name]

	if !ok {
		return nil, errors.New("undefined: " + name)
	}
	return val, nil
}

type Evaluator struct {
	actions map[string]Action
	symtab  symtab
}

func (e *Evaluator) AddAction(a Action) {
	if e.actions == nil {
		e.actions = make(map[string]Action)
	}
	e.actions[a.Name] = a
}

func (e *Evaluator) interpolate(n syntax.Node, s string) (Value, error) {
	return String{Value: s}, nil
}

func (e *Evaluator) doEval(n syntax.Node) (Value, error) {
	switch v := n.(type) {
	case *syntax.VarDecl:
		val, err := e.doEval(v.Value)

		if err != nil {
			return nil, v.Err(err.Error())
		}
		e.symtab.put(v.Ident.Name, val)
	case *syntax.Ref:
		switch v := v.Left.(type) {
		case *syntax.Ident:
			return e.symtab.get(v.Name)
		case *syntax.DotExpr:
		case *syntax.IndExpr:
		default:
			return nil, v.Err("invalid reference")
		}
	case *syntax.Lit:
		switch v.Type {
		case token.String:
			return e.interpolate(v, v.Value)
		case token.Int:
			i, _ := strconv.ParseInt(v.Value, 10, 64)
			return Int{Value: i}, nil
		case token.Bool:
			b := true

			if v.Value != "true" {
				b = false
			}
			return Bool{Value: b}, nil
		}
	case *syntax.Ident:
	case *syntax.Array:
		items := make([]Value, 0, len(v.Items))

		for _, it := range v.Items {
			val, err := e.doEval(it)

			if err != nil {
				return nil, it.Err(err.Error())
			}
			items = append(items, val)
		}
		return Array{Items: items}, nil
	case *syntax.Object:
		pairs := make(map[Key]Value)

		for _, n := range v.Pairs {
			val, err := e.doEval(n.Value)

			if err != nil {
				return nil, n.Value.Err(err.Error())
			}
			pairs[Key{Name: n.Key.Name}] = val
		}
		return Object{Pairs: pairs}, nil
	case *syntax.BlockStmt:
		for _, n := range v.Nodes {
			val, err := e.doEval(n)

			if err != nil {
				return nil, n.Err(err.Error())
			}

			if _, ok := val.(Yield); ok {
				return val, nil
			}
		}
		return nil, nil
	case *syntax.ActionStmt:
		action, ok := e.actions[v.Name]

		if !ok {
			return nil, v.Err("undefined action: " + v.Name)
		}

		args := make([]Value, 0, len(v.Args))

		for _, n := range v.Args {
			val, err := e.doEval(n)

			if err != nil {
				return nil, n.Err(err.Error())
			}
			args = append(args, val)
		}

		var dest Value

		if v.Dest != nil {
			val, err := e.doEval(v.Dest)

			if err != nil {
				return nil, v.Dest.Err(err.Error())
			}
			dest = val
		}
		return action.Call(args, dest)
	case *syntax.YieldStmt:
		val, err := e.doEval(v.Value)

		if err != nil {
			return nil, v.Err(err.Error())
		}
		return Yield{Value: val}, nil
	case *syntax.IfStmt:
	}
	return nil, nil
}

func (e *Evaluator) Eval(nn []syntax.Node) error {
	for _, n := range nn {
		if _, err := e.doEval(n); err != nil {
			return err
		}
	}
	return nil
}
