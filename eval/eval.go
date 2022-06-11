// Package eval handles the Evaluation of a req script.
package eval

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"strconv"
	"unicode/utf8"
	"github.com/andrewpillar/req/syntax"
	"github.com/andrewpillar/req/value"
)

// Context stores the variables that have been set during a script's Evaluation.
type Context struct {
	symtab map[string]value.Value
}

// Put puts the given valect into the current Context under the given name.
func (c *Context) Put(name string, val value.Value) {
	if c.symtab == nil {
		c.symtab = make(map[string]value.Value)
	}
	c.symtab[name] = val
}

func errUndefined(name string) error {
	return errors.New("undefined: " + name)
}

// Get returns an valect of the given name. If no valect exists, then this
// errors.
func (c *Context) Get(name string) (value.Value, error) {
	if c.symtab == nil {
		return nil, errUndefined(name)
	}

	val, ok := c.symtab[name]

	if !ok {
		return nil, errUndefined(name)
	}
	return val, nil
}

// Copy returns a copy of the current Context.
func (c *Context) Copy() *Context {
	c2 := &Context{
		symtab: make(map[string]value.Value),
	}

	for k, v := range c.symtab {
		c2.symtab[k] = v
	}
	return c2
}

// Error records an error that occurred during Evaluation and the position at
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
	// during Evaluation such as file handles. These are not called if the
	// "exit" command is called however.
	finalizers []func() error
}

var builtinCmds = []*Command{
	CookieCmd,
	DecodeCmd,
	EncodeCmd,
	EnvCmd,
	ExitCmd,
	OpenCmd,
	ReadCmd,
	ReadlnCmd,
	WriteCmd,
	WritelnCmd,
	HeadCmd,
	OptionsCmd,
	GetCmd,
	PutCmd,
	PostCmd,
	PatchCmd,
	DeleteCmd,
	TlsCmd,
	SendCmd,
	SniffCmd,
}

// New returns a new evaluator for evaluating req scripts. The given writer is
// used as the standard output for the write and writeln commands.
func New(out io.Writer) *Evaluator {
	e := &Evaluator{
		cmds: make(map[string]*Command),
	}

	WriteCmd.Func = write(out)
	WritelnCmd.Func = writeln(out)

	for _, cmd := range builtinCmds {
		e.AddCmd(cmd)
	}
	return e
}

// AddCmd adds the given command to the Evaluator.
func (e *Evaluator) AddCmd(cmd *Command) {
	if e.cmds == nil {
		e.cmds = make(map[string]*Command)
	}
	e.cmds[cmd.Name] = cmd
}

// interpolate parses the given string for $(Ref), $(Ref.Dot), and $(Ref[Ind])
// expressions and interpolates any that are found using the given Context.
func (e *Evaluator) interpolate(c *Context, litpos syntax.Pos, s string) (value.Value, error) {
	var buf bytes.Buffer

	interpolate := false
	expr := make([]rune, 0, len(s))

	pos := litpos
	end := len(s) - 1

	i := 0
	w := 1

	for i <= end {
		r := rune(s[i])

		if r >= utf8.RuneSelf {
			r, w = utf8.DecodeRune([]byte(s[i:]))
		}

		i += w

		if r == '\\' {
			if i <= end {
				switch s[i] {
				case 't':
					buf.WriteRune('\t')
					i++
					continue
				case 'r':
					buf.WriteRune('\r')
					i++
					continue
				case 'n':
					buf.WriteRune('\n')
					i++
					continue
				}
				continue
			}
		}

		if r == '$' {
			if i <= end && s[i] == '(' {
				interpolate = true
				pos.Col += i
				i++ // skip ahead beyond the opening (
				continue
			}
		}

		if r == ')' && interpolate {
			interpolate = false

			n, err := syntax.ParseRef("$" + string(expr))

			if err != nil {
				return nil, Error{
					Pos: pos,
					Err: err,
				}
			}

			val, err := e.Eval(c, n)

			if err != nil {
				return nil, Error{
					Pos: pos,
					Err: errors.Unwrap(err),
				}
			}

			buf.WriteString(val.Sprint())
			expr = expr[0:0]
			pos.Col = 0
			continue
		}

		if interpolate {
			expr = append(expr, r)
			continue
		}
		buf.WriteRune(r)
	}

	return value.String{
		Value: buf.String(),
	}, nil
}

// resolveCommand resolves the given command node into a command and its
// arguments that can be used for command invocation.
func (e *Evaluator) resolveCommand(c *Context, n *syntax.CommandStmt) (*Command, []value.Value, error) {
	cmd, ok := e.cmds[n.Name.Value]

	if !ok {
		return nil, nil, errors.New("undefined command: " + n.Name.Value)
	}

	args := make([]value.Value, 0, len(n.Args))

	for _, arg := range n.Args {
		val, err := e.Eval(c, arg)

		if err != nil {
			return nil, nil, e.err(arg.Pos(), err)
		}
		args = append(args, val)
	}
	return cmd, args, nil
}

// resolveDot resolves the given dot expression with the given Context and
// returns the valect that is being referred to via the expression if any.
func (e *Evaluator) resolveDot(c *Context, n *syntax.DotExpr) (value.Value, error) {
	left, err := e.Eval(c, n.Left)

	if err != nil {
		return nil, err
	}

	val := left

	if name, err := value.ToName(left); err == nil {
		val, err = c.Get(name.Value)

		if err != nil {
			return nil, err
		}
	}

	sel, err := value.ToSelector(val)

	if err != nil {
		return nil, err
	}

	right, err := e.Eval(c, n.Right)

	if err != nil {
		return nil, err
	}

	val, err = sel.Select(right)

	if err != nil {
		return nil, err
	}
	return val, nil
}

// resolveIndex resolves the given index expression with the given Context and
// returns the valect that is being referred to via the expression if any.
func (e *Evaluator) resolveIndex(c *Context, n *syntax.IndExpr) (value.Value, error) {
	left, err := e.Eval(c, n.Left)

	if err != nil {
		return nil, err
	}

	if name, ok := left.(value.Name); ok {
		left, err = c.Get(name.Value)

		if err != nil {
			return nil, err
		}
	}

	index, err := value.ToIndex(left)

	if err != nil {
		return nil, err
	}

	right, err := e.Eval(c, n.Right)

	if err != nil {
		return nil, err
	}
	return index.Get(right)
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

type branchErr struct {
	kind string
	pos  syntax.Pos
}

func (e branchErr) Error() string {
	return e.pos.String() + " - " + e.kind + " outside of loop"
}

// evalAssign evaluates the node and assigns the given value to that node. If
// the given node is a Name then it simply assigns the value directly to the
// Name's value in the symbol table. If the node is an IndExpr then the
// expression is evaluated to find the index being assigned to, and evaluates
// the key in the index to where the value is to be assigned.
func (e *Evaluator) evalAssign(c *Context, strict bool, n syntax.Node, val value.Value) error {
	switch v := n.(type) {
	case *syntax.Name:
		if v.Value == "_" {
			return nil
		}

		orig, _ := c.Get(v.Value)

		if strict {
			if orig != nil {
				if err := value.CompareType(val, orig); err != nil {
					return err
				}
			}
		}

		c.Put(v.Value, val)
		return nil
	case *syntax.IndExpr:
		left, err := e.Eval(c, v.Left)

		if err != nil {
			return err
		}

		if name, ok := left.(value.Name); ok {
			left, err = c.Get(name.Value)

			if err != nil {
				return err
			}
		}

		index, err := value.ToIndex(left)

		if err != nil {
			return err
		}

		key, err := e.Eval(c, v.Right)

		if err != nil {
			return err
		}

		if err := index.Set(strict, key, val); err != nil {
			return err
		}
		return nil
	}
	return errors.New("unexpected expression")
}

func (e *Evaluator) evalRange(c *Context, n *syntax.Range, body *syntax.BlockStmt) (value.Value, error) {
	right, err := e.Eval(c, n.Right)

	if err != nil {
		return nil, err
	}

	iter, err := value.ToIterable(right)

	if err != nil {
		return nil, e.err(n.Right.Pos(), err)
	}

	list, ok := n.Left.(*syntax.ExprList)

	if !ok {
		return nil, e.err(n.Pos(), errors.New("assignment is not to a list of variables"))
	}

	l := len(list.Nodes)

	if l > 2 {
		return nil, e.err(n.Pos(), errors.New("assignment mismatch: can only assign at most 2 variables during iteration"))
	}

	key, val, err := iter.Next()

loop:
	for !errors.Is(err, io.EOF) {
		if l >= 1 {
			if err := e.evalAssign(c, false, list.Nodes[0], key); err != nil {
				return nil, e.err(n.Pos(), err)
			}

			if l > 1 {
				if err := e.evalAssign(c, false, list.Nodes[1], val); err != nil {
					return nil, e.err(list.Nodes[1].Pos(), err)
				}
			}
		}

		if _, err := e.Eval(c, body); err != nil {
			if branch, ok := err.(branchErr); ok {
				switch branch.kind {
				case "break":
					break loop
				case "continue":
					goto cont
				}
			}
			return nil, e.err(body.Pos(), err)
		}

	cont:
		key, val, err = iter.Next()
	}
	return nil, nil
}

// Eval Evaluates the given node and returns the value it Evaluates to if any.
func (e *Evaluator) Eval(c *Context, n syntax.Node) (value.Value, error) {
	switch v := n.(type) {
	case *syntax.AssignStmt:
		list, ok := v.Left.(*syntax.ExprList)

		if !ok {
			return nil, e.err(v.Left.Pos(), errors.New("assignment is not to a list of variables"))
		}

		right, ok := v.Right.(*syntax.ExprList)

		if !ok {
			return nil, e.err(v.Right.Pos(), errors.New("assignment is not from a list of expressions"))
		}

		if len(list.Nodes) != len(right.Nodes) {
			return nil, e.err(v.Pos(), fmt.Errorf("assignment mismatch: %d variable(s) but %d value(s)", len(right.Nodes), len(list.Nodes)))
		}

		for i, n := range list.Nodes {
			valnod := right.Nodes[i]

			val, err := e.Eval(c, valnod)

			if err != nil {
				return nil, e.err(valnod.Pos(), err)
			}

			if err := e.evalAssign(c, true, n, val); err != nil {
				return nil, e.err(v.Pos(), err)
			}
		}
	case *syntax.Ref:
		switch v := v.Left.(type) {
		case *syntax.Name:
			val, err := c.Get(v.Value)

			if err != nil {
				return nil, e.err(v.Pos(), err)
			}
			return val, nil
		case *syntax.DotExpr:
			val, err := e.resolveDot(c, v)

			if err != nil {
				return nil, e.err(v.Pos(), err)
			}
			return val, nil
		case *syntax.IndExpr:
			val, err := e.resolveIndex(c, v)

			if err != nil {
				return nil, e.err(v.Pos(), err)
			}
			return val, nil
		default:
			return nil, e.err(v.Pos(), errors.New("invalid reference"))
		}
	case *syntax.DotExpr:
		val, err := e.resolveDot(c, v)

		if err != nil {
			return nil, e.err(v.Pos(), err)
		}
		return val, nil
	case *syntax.IndExpr:
		val, err := e.resolveIndex(c, v)

		if err != nil {
			return nil, e.err(v.Pos(), err)
		}
		return val, nil
	case *syntax.Lit:
		switch v.Type {
		case syntax.StringLit:
			val, err := e.interpolate(c, v.Pos(), v.Value)

			if err != nil {
				return nil, err
			}
			return val, err
		case syntax.IntLit:
			i, _ := strconv.ParseInt(v.Value, 10, 64)
			return value.Int{Value: i}, nil
		case syntax.FloatLit:
			f, _ := strconv.ParseFloat(v.Value, 64)
			return value.Float{Value: f}, nil
		case syntax.BoolLit:
			b := true

			if v.Value != "true" {
				b = false
			}
			return value.Bool{Value: b}, nil
		}
	case *syntax.Name:
		return value.Name{Value: v.Value}, nil
	case *syntax.Array:
		items := make([]value.Value, 0, len(v.Items))

		for _, it := range v.Items {
			val, err := e.Eval(c, it)

			if err != nil {
				return nil, e.err(it.Pos(), err)
			}
			items = append(items, val)
		}

		arr, err := value.NewArray(items)

		if err != nil {
			return nil, e.err(v.Pos(), err)
		}
		return arr, nil
	case *syntax.Object:
		pairs := make(map[string]value.Value)
		order := make([]string, 0, len(v.Pairs))

		for _, n := range v.Pairs {
			val, err := e.Eval(c, n.Value)

			if err != nil {
				return nil, err
			}
			pairs[n.Key.Value] = val
			order = append(order, n.Key.Value)
		}

		return &value.Object{
			Order: order,
			Pairs: pairs,
		}, nil
	case *syntax.BlockStmt:
		// Create a copy so we can unset any variables that will fall out of
		// scope of the block.
		orig := c.Copy()

		for _, n := range v.Nodes {
			if _, err := e.Eval(c, n); err != nil {
				return nil, err
			}
		}

		// Delete any variables that do not exist in the original context.
		for name := range c.symtab {
			if _, ok := orig.symtab[name]; !ok {
				delete(c.symtab, name)
			}
		}
		return nil, nil
	case *syntax.CommandStmt:
		cmd, args, err := e.resolveCommand(c, v)

		if err != nil {
			return nil, e.err(n.Pos(), err)
		}

		val, err := cmd.invoke(args)

		if err != nil {
			return nil, err
		}

		if f, ok := val.(value.File); ok {
			e.finalizers = append(e.finalizers, f.Close)
		}
		return val, nil
	case *syntax.MatchStmt:
		condval, err := e.Eval(c, v.Cond)

		if err != nil {
			return nil, err
		}

		jmptab := make(map[uint32]syntax.Node)

		for _, stmt := range v.Cases {
			h := fnv.New32a()

			val, err := e.Eval(c, stmt.Value)

			if err != nil {
				return nil, err
			}

			if err := value.CompareType(condval, val); err != nil {
				return nil, e.err(stmt.Pos(), err)
			}

			h.Write([]byte(val.String()))

			jmptab[h.Sum32()] = stmt.Then
		}

		h := fnv.New32a()
		h.Write([]byte(condval.String()))

		if n, ok := jmptab[h.Sum32()]; ok {
			return e.Eval(c, n)
		}

		if v.Default != nil {
			return e.Eval(c, v.Default)
		}
		return nil, nil
	case *syntax.ChainExpr:
		var val value.Value

		for _, n := range v.Commands {
			cmd, args, err := e.resolveCommand(c, n)

			if err != nil {
				return nil, e.err(n.Pos(), err)
			}

			if val != nil {
				args = append(args, val)
			}

			val, err = cmd.invoke(args)

			if err != nil {
				return nil, err
			}
		}
		return val, nil
	case *syntax.IfStmt:
		val, err := e.Eval(c, v.Cond)

		if err != nil {
			return nil, err
		}

		if value.Truthy(val) {
			return e.Eval(c, v.Then)
		}

		if v.Else != nil {
			return e.Eval(c, v.Else)
		}
		return nil, nil
	case *syntax.Operation:
		left, err := e.Eval(c, v.Left)

		if err != nil {
			return nil, err
		}

		if v.Right == nil {
			return value.Bool{
				Value: value.Truthy(left),
			}, nil
		}

		right, err := e.Eval(c, v.Right)

		if err != nil {
			return nil, e.err(v.Right.Pos(), err)
		}

		val, err := value.Compare(left, v.Op, right)

		if err != nil {
			return nil, e.err(v.Pos(), err)
		}
		return val, nil
	case *syntax.ForStmt:
		c2 := c.Copy()

		if v.Init != nil {
			if rng, ok := v.Init.(*syntax.Range); ok {
				return e.evalRange(c2, rng, v.Body)
			}

			if _, err := e.Eval(c2, v.Init); err != nil {
				return nil, e.err(v.Pos(), err)
			}
		}

	loop:
		for {
			if v.Cond != nil {
				val, err := e.Eval(c2, v.Cond)

				if err != nil {
					return nil, e.err(v.Pos(), err)
				}

				if !value.Truthy(val) {
					break
				}
			}

			if _, err := e.Eval(c2, v.Body); err != nil {
				// Feels like a hack but we'll see...
				if branch, ok := err.(branchErr); ok {
					switch branch.kind {
					case "break":
						break loop
					case "continue":
						goto cont
					}
				}
				return nil, e.err(v.Body.Pos(), err)
			}

		cont:
			if v.Post != nil {
				if _, err := e.Eval(c2, v.Post); err != nil {
					return nil, e.err(v.Pos(), err)
				}
			}
		}
	case *syntax.BranchStmt:
		return nil, branchErr{kind: v.Tok.String(), pos: v.Pos()}
	}
	return nil, nil
}

// Run Evaluates all of the given nodes.
func (e *Evaluator) Run(nn []syntax.Node) error {
	var c Context

	for _, n := range nn {
		if _, err := e.Eval(&c, n); err != nil {
			return err
		}
	}

	for _, fn := range e.finalizers {
		fn()
	}
	return nil
}
