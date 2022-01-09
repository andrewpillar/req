// Package eval handles the evaluation of a req script.
package eval

import (
	"bytes"
	"errors"
	"hash/fnv"
	"strconv"

	"github.com/andrewpillar/req/syntax"
)

// context stores the variables that have been set during a script's evaluation.
type context struct {
	symtab map[string]Object
}

// Put puts the given object into the current context under the given name.
func (c *context) Put(name string, obj Object) {
	if c.symtab == nil {
		c.symtab = make(map[string]Object)
	}
	c.symtab[name] = obj
}

type errUndefined struct {
	name string
}

func (e errUndefined) Error() string { return "undefined: $" + e.name }

// Get returns an object of the given name. If no object exists, then this
// errors.
func (c *context) Get(name string) (Object, error) {
	if c.symtab == nil {
		return nil, errUndefined{name: name}
	}

	obj, ok := c.symtab[name]

	if !ok {
		return nil, errUndefined{name: name}
	}
	return obj, nil
}

// Copy returns a copy of the current context.
func (c *context) Copy() *context {
	c2 := &context{
		symtab: make(map[string]Object),
	}

	for k, v := range c.symtab {
		c2.symtab[k] = v
	}
	return c2
}

// Error records an error that occurred during evaluation and the position at
// which the error occurred and the original error itself.
type Error struct {
	Pos syntax.Pos
	Err error
}

func (e Error) Unwrap() error { return e.Err }
func (e Error) Error() string { return e.Pos.String() + " - " + e.Err.Error() }

type Evaluator struct {
	cmds map[string]*Command

	// slice of cleanup functions to call to cleanup any resources opened
	// during evaluation such as file handles. These are not called if the
	// "exit" command is called however.
	finalizers []func() error
}

var builtinCmds = []*Command{
	EnvCmd,
	ExitCmd,
	OpenCmd,
	PrintCmd,
	HeadCmd,
	OptionsCmd,
	GetCmd,
	PutCmd,
	PostCmd,
	PatchCmd,
	DeleteCmd,
	SendCmd,
	SniffCmd,
}

func New() *Evaluator {
	e := &Evaluator{
		cmds: make(map[string]*Command),
	}

	for _, cmd := range builtinCmds {
		e.AddCmd(cmd)
	}
	return e
}

// AddCmd adds the given command to the evaluator.
func (e *Evaluator) AddCmd(cmd *Command) {
	if e.cmds == nil {
		e.cmds = make(map[string]*Command)
	}
	e.cmds[cmd.Name] = cmd
}

// interpolate parses the given string for {$Ref}, {$Ref.Dot}, and {$Ref[Ind]}
// expressions and interpolates any that are found using the given context.
func (e *Evaluator) interpolate(c *context, s string) (Object, error) {
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
				return nil, err
			}

			obj, err := e.eval(c, n)

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

// resolveCommand resolves the given command node into a command and its
// arguments that can be used for command invocation.
func (e *Evaluator) resolveCommand(c *context, n *syntax.CommandStmt) (*Command, []Object, error) {
	cmd, ok := e.cmds[n.Name.Value]

	if !ok {
		return nil, nil, errors.New("undefined command: " + n.Name.Value)
	}

	args := make([]Object, 0, len(n.Args))

	for _, arg := range n.Args {
		obj, err := e.eval(c, arg)

		if err != nil {
			return nil, nil, e.err(arg.Pos(), err)
		}
		args = append(args, obj)
	}
	return cmd, args, nil
}

// resolveArrayIndex returns the object in the given array at the given index
// if any. If there is no object, then nil is returned.
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
		return zeroObj{}, nil
	}
	return arrobj.items[i], nil
}

// resolveHashKey returns the object in the given hash under the given key if
// any. If there is no object, then nil is returned.
func (e *Evaluator) resolveHashKey(hash, key Object) (Object, error) {
	s, ok := key.(stringObj)

	if !ok {
		return nil, TypeError{
			typ:      key.Type(),
			expected: String,
		}
	}

	hashobj := hash.(hashObj)

	obj, ok := hashobj.pairs[s.value]

	if !ok {
		return zeroObj{}, nil
	}
	return obj, nil
}

// resolveDot resolves the given dot expression with the given context and
// returns the object that is being referred to via the expression if any.
func (e *Evaluator) resolveDot(c *context, n *syntax.DotExpr) (Object, error) {
	left, err := e.eval(c, n.Left)

	if err != nil {
		return nil, err
	}

	name, ok := left.(nameObj)

	if !ok {
		return nil, errors.New("cannot use type " + left.Type().String() + " as selector")
	}

	obj, err := c.Get(name.value)

	if err != nil {
		return nil, err
	}

	sel, ok := obj.(Selector)

	if !ok {
		return nil, errors.New("cannot use type " + obj.Type().String() + " as selector")
	}

	right, err := e.eval(c, n.Right)

	if err != nil {
		return nil, err
	}

	obj, err = sel.Select(right)

	if err != nil {
		return nil, err
	}
	return obj, nil
}

// resolveInd resolves the given index expression with the given context and
// returns the object that is being referred to via the expression if any.
func (e *Evaluator) resolveInd(c *context, n *syntax.IndExpr) (Object, error) {
	left, err := e.eval(c, n.Left)

	if err != nil {
		return nil, err
	}

	var obj Object

	switch v := left.(type) {
	case nameObj:
		obj, err = c.Get(v.value)

		if err != nil {
			return nil, err
		}
	case arrayObj:
		obj = v
	case hashObj:
		obj = v
	default:
		return nil, errors.New("type " + left.Type().String() + " does not support indexing")
	}

	right, err := e.eval(c, n.Right)

	if err != nil {
		return nil, err
	}

	switch obj.Type() {
	case Array:
		return e.resolveArrayIndex(obj, right)
	case Hash:
		return e.resolveHashKey(obj, right)
	default:
		return nil, errors.New("type " + obj.Type().String() + " does not support indexing")
	}
}

// err records the given error at the given position. If the given error is of
// type Error then no record is made, this is to prevent superfluous recording
// of position information.
func (e *Evaluator) err(pos syntax.Pos, err error) error {
	if _, ok := err.(Error); ok {
		return err
	}

	return Error{
		Pos: pos,
		Err: err,
	}
}

// eval evaluates the given node with the given context and returns the object
// the node evaluates to, if any.
func (e *Evaluator) eval(c *context, n syntax.Node) (Object, error) {
	switch v := n.(type) {
	case *syntax.VarDecl:
		obj, err := e.eval(c, v.Value)

		if err != nil {
			return nil, e.err(v.Value.Pos(), err)
		}

		if obj == nil {
			return nil, e.err(v.Value.Pos(), errors.New("does not evaluate to value"))
		}
		c.Put(v.Name.Value, obj)
	case *syntax.Ref:
		switch v := v.Left.(type) {
		case *syntax.Name:
			obj, err := c.Get(v.Value)

			if err != nil {
				return nil, e.err(v.Pos(), err)
			}
			return obj, nil
		case *syntax.DotExpr:
			obj, err := e.resolveDot(c, v)

			if err != nil {
				return nil, e.err(v.Pos(), err)
			}
			return obj, nil
		case *syntax.IndExpr:
			obj, err := e.resolveInd(c, v)

			if err != nil {
				return nil, e.err(v.Pos(), err)
			}
			return obj, nil
		default:
			return nil, errors.New("invalid reference")
		}
	case *syntax.DotExpr:
		obj, err := e.resolveDot(c, v)

		if err != nil {
			return nil, e.err(v.Pos(), err)
		}
		return obj, nil
	case *syntax.IndExpr:
		obj, err := e.resolveInd(c, v)

		if err != nil {
			return nil, e.err(v.Pos(), err)
		}
		return obj, nil
	case *syntax.Lit:
		switch v.Type {
		case syntax.StringLit:
			obj, err := e.interpolate(c, v.Value)

			if err != nil {
				// Offset original position of string so we report the position
				// in the evaluated expression.
				evalerr := err.(Error)
				pos := v.Pos()
				pos.Col += evalerr.Pos.Col + 1

				return nil, e.err(pos, evalerr.Err)
			}
			return obj, err
		case syntax.IntLit:
			i, _ := strconv.ParseInt(v.Value, 10, 64)
			return intObj{value: i}, nil
		case syntax.BoolLit:
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
			obj, err := e.eval(c, it)

			if err != nil {
				return nil, err
			}
			items = append(items, obj)
		}
		return arrayObj{
			items: items,
		}, nil
	case *syntax.Object:
		pairs := make(map[string]Object)

		for _, n := range v.Pairs {
			obj, err := e.eval(c, n.Value)

			if err != nil {
				return nil, err
			}
			pairs[n.Key.Value] = obj
		}
		return hashObj{pairs: pairs}, nil
	case *syntax.BlockStmt:
		// Make a copy of the current context so we can correctly shadow any
		// variables declared outside of the block.
		c2 := c.Copy()

		for _, n := range v.Nodes {
			if _, err := e.eval(c2, n); err != nil {
				return nil, err
			}
		}
		return nil, nil
	case *syntax.CommandStmt:
		cmd, args, err := e.resolveCommand(c, v)

		if err != nil {
			return nil, e.err(n.Pos(), err)
		}

		obj, err := cmd.invoke(args)

		if err != nil {
			return nil, err
		}

		if f, ok := obj.(fileObj); ok {
			e.finalizers = append(e.finalizers, f.Close)
		}
		return obj, nil
	case *syntax.MatchStmt:
		obj, err := e.eval(c, v.Cond)

		if err != nil {
			return nil, err
		}

		jmptab := make(map[uint32]syntax.Node)

		for _, stmt := range v.Cases {
			h := fnv.New32a()

			obj, err := e.eval(c, stmt.Value)

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
			return e.eval(c, n)
		}

		if v.Default != nil {
			return e.eval(c, v.Default)
		}
		return nil, nil
	case *syntax.ChainExpr:
		var obj Object

		for _, n := range v.Commands {
			cmd, args, err := e.resolveCommand(c, n)

			if err != nil {
				return nil, err
			}

			if obj != nil {
				args = append([]Object{obj}, args...)
			}

			obj, err = cmd.invoke(args)

			if err != nil {
				return nil, err
			}
		}
		return obj, nil
	case *syntax.IfStmt:
	}
	return nil, nil
}

// Run evaluates all of the given nodes.
func (e *Evaluator) Run(nn []syntax.Node) error {
	var c context

	for _, n := range nn {
		if _, err := e.eval(&c, n); err != nil {
			return err
		}
	}

	for _, fn := range e.finalizers {
		fn()
	}
	return nil
}
