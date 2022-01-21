// Package syntax provides functions for parsing req scripts into their ASTs
// for evaluation.
package syntax

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type parser struct {
	*scanner

	errc int
}

// ParseExpr parses all of the expressions from the given string. This is would
// be used as part of a REPL to parse each line that is input.
func ParseExpr(s string) ([]Node, error) {
	errs := make([]error, 0)

	p := parser{
		scanner: newScanner(newSource("", strings.NewReader(s), func(pos Pos, msg string) {
			errs = append(errs, errors.New(msg))
		})),
	}

	nn, err := p.parse(true)

	if err != nil {
		return nil, errs[0]
	}
	return nn, nil
}

// ParseRef is a convenience function for parsing a single $Ref, $Ref.Dot,
// or $Ref[Ind] expression. This is used as part of string interpolation. If
// multiple errors occur during parsing, then the first of these errors is
// returned.
func ParseRef(s string) (Node, error) {
	errs := make([]error, 0)

	p := parser{
		scanner: newScanner(newSource("", strings.NewReader(s), func(pos Pos, msg string) {
			errs = append(errs, errors.New(msg))
		})),
	}

	if p.tok != _Ref {
		return nil, errors.New("expected $")
	}

	n := p.ref()

	if len(errs) > 0 {
		return nil, errs[0]
	}
	return n, nil
}

// Parse reads all content from the given reader and parses it into an AST. The
// given callback is used to reporting any and all errors that may occur during
// parsing, using the given name for uniquely identifying the parser (typically
// a filename).
func Parse(name string, r io.Reader, errh func(Pos, string)) ([]Node, error) {
	p := parser{
		scanner: newScanner(newSource(name, r, errh)),
	}
	return p.parse(false)
}

// ParseFile is a convenience function that parses the given file useing Parse.
func ParseFile(fname string, errh func(Pos, string)) ([]Node, error) {
	f, err := os.Open(fname)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	return Parse(fname, f, errh)
}

// advance moves the parser along the given follow set of tokens and stops
// when it encounters the first one. This will always advance to EOF if none
// of the tokens can be found.
func (p *parser) advance(follow ...token) {
	set := make(map[token]struct{})

	for _, tok := range follow {
		set[tok] = struct{}{}
	}
	set[_EOF] = struct{}{}

	for {
		if _, ok := set[p.tok]; ok {
			break
		}
		p.next()
	}
}

// got will consume the given token if it matches what we currently have, and
// returns whether or not it was consumed.
func (p *parser) got(tok token) bool {
	if p.tok == tok {
		p.next()
		return true
	}
	return false
}

// errAt reports an error at the given position.
func (p *parser) errAt(pos Pos, msg string) {
	p.errc++
	p.scanner.source.errh(pos, msg)
}

// err reports an error at the current position.
func (p *parser) err(msg string) {
	p.errAt(p.pos, msg)
}

func (p *parser) expected(tok token) {
	p.err("expected " + tok.String())
}

func (p *parser) unexpected(tok token) {
	p.err("unexpected " + tok.String())
}

// want will attempt to consume the given token. If the given token cannot be
// consumed, then it reports an error.
func (p *parser) want(tok token) {
	if !p.got(tok) {
		p.expected(tok)
	}
}

// node returns a new node at the current position for use in constructing
// a valid node in the AST.
func (p *parser) node() node {
	return node{
		pos: p.pos,
	}
}

func (p *parser) name() *Name {
	if p.tok != _Name {
		return nil
	}

	n := &Name{
		node:  p.node(),
		Value: p.lit,
	}

	p.next()
	return n
}

// nameExpr will parse either a name expression, or an index expression. This
// would be used for the left-hand side of an assignment statement, where you
// want to assign to an indexed value, for example Arr[0] = "val".
func (p *parser) nameExpr() Node {
	name := p.name()

	if p.tok != _Lbrack {
		return name
	}

	var n Node = name

	for {
		pos := p.pos

		if p.tok != _Lbrack {
			break
		}

		p.next()

		ind := &IndExpr{
			node: node{pos: pos},
			Left: n,
		}

		switch p.tok {
		case _Literal:
			ind.Right = p.literal()
		case _Ref:
			ind.Right = p.ref()
		case _Rbrack:
			ind.Right = &Array{}
		default:
			p.unexpected(p.tok)
			p.next()
		}

		p.want(_Rbrack)

		n = ind
	}
	return n
}

func (p *parser) literal() *Lit {
	if p.tok != _Literal {
		return nil
	}

	n := &Lit{
		node:  p.node(),
		Type:  p.typ,
		Value: p.lit,
	}
	p.next()
	return n
}

// ref parses a variable reference expression. This will parse the $Ref,
// $Left.Right, and $Left[Right] expressions recursively.
func (p *parser) ref() *Ref {
	if p.tok != _Ref {
		return nil
	}

	p.got(_Ref)

	if p.tok != _Name {
		p.expected(_Name)
	}

	ref := &Ref{
		node: p.node(),
		Left: p.name(),
	}

loop:
	for {
		pos := p.pos

		switch p.tok {
		case _Dot:
			p.next()

			if p.tok != _Name {
				p.expected(_Name)
				p.next()
				return nil
			}

			left := ref.Left

			ref.Left = &DotExpr{
				node:  node{pos: pos},
				Left:  left,
				Right: p.name(),
			}
		case _Lbrack:
			p.next()

			if p.tok == _Rbrack {
				p.err("expected string, int, or variable")
				p.next()
				break
			}

			left := ref.Left
			ind := &IndExpr{
				node: node{pos: pos},
				Left: left,
			}

			switch p.tok {
			case _Literal:
				ind.Right = p.literal()
			case _Ref:
				ind.Right = p.ref()
			default:
				p.unexpected(p.tok)
				p.next()
			}

			p.want(_Rbrack)
			ref.Left = ind
		default:
			break loop
		}
	}
	return ref
}

// list parses all of the tokens in a list with the given separator of sep, and
// end token of end. The given callback parse is called to actually handle the
// parsing of tokens.
func (p *parser) list(sep, end token, parse func()) {
	for p.tok != _EOF && p.tok != end {
		parse()

		if !p.got(sep) && p.tok != end {
			p.err("expected " + sep.String() + " or " + end.String())
			p.next()
		}
	}
	p.want(end)
}

func (p *parser) obj() *Object {
	p.want(_Lparen)

	n := &Object{
		node: p.node(),
	}

	p.list(_Comma, _Rparen, func() {
		if p.tok != _Name {
			p.expected(_Name)
			p.advance(_Rparen, _Semi)
			return
		}

		key := p.name()

		p.want(_Colon)

		n.Pairs = append(n.Pairs, &KeyExpr{
			node:  p.node(),
			Key:   key,
			Value: p.expr(),
		})
	})
	return n
}

func (p *parser) arr() *Array {
	p.want(_Lbrack)

	n := &Array{
		node: p.node(),
	}

	p.list(_Comma, _Rbrack, func() {
		n.Items = append(n.Items, p.operand())
	})
	return n
}

func (p *parser) blockstmt() *BlockStmt {
	p.want(_Lbrace)

	n := &BlockStmt{
		node: p.node(),
	}

	for p.tok != _Rbrace && p.tok != _EOF {
		n.Nodes = append(n.Nodes, p.stmt(false))
	}

	p.want(_Rbrace)
	return n
}

func (p *parser) casestmt() *CaseStmt {
	n := &CaseStmt{
		node: p.node(),
	}

	if p.tok != _Literal {
		p.unexpected(p.tok)
		p.next()
		return nil
	}

	n.Value = p.literal()

	p.want(_Arrow)

	switch p.tok {
	case _Lbrace:
		n.Then = p.blockstmt()
	case _Name:
		n.Then = p.command(p.name())
	default:
		p.unexpected(p.tok)
		p.next()
	}
	return n
}

func (p *parser) matchstmt() *MatchStmt {
	if p.tok != _Match {
		return nil
	}

	n := &MatchStmt{
		node: p.node(),
	}

	p.next()

	switch p.tok {
	case _Literal:
		n.Cond = p.literal()
	case _Ref:
		n.Cond = p.ref()
	default:
		p.unexpected(p.tok)
		p.next()
	}

	p.want(_Lbrace)

	for p.tok != _Rbrace {
		if p.tok == _Name {
			if p.lit != "_" {
				p.unexpected(_Name)
				p.advance(_Rbrace, _Semi)
				continue
			}

			p.name()
			p.want(_Arrow)

			switch p.tok {
			case _Lbrace:
				n.Default = p.blockstmt()
			case _Name:
				n.Default = p.command(p.name())
			default:
				p.unexpected(p.tok)
				p.next()
			}
			p.got(_Semi)
			continue
		}

		n.Cases = append(n.Cases, p.casestmt())

		p.got(_Semi)
	}

	p.got(_Rbrace)
	return n
}

func (p *parser) ifstmt() *IfStmt {
	nodpos := p.node()

	if !p.got(_If) {
		return nil
	}

	n := &IfStmt{
		node: nodpos,
		Cond: p.expr(),
	}

	if p.tok != _Lbrace {
		p.errAt(n.Pos(), "missing condition in if statement")
		return n
	}

	n.Then = p.blockstmt()

	if p.got(_Else) {
		switch p.tok {
		case _If:
			n.Else = p.ifstmt()
		case _Lbrace:
			n.Else = p.blockstmt()
		default:
			p.err("expected if statement or {")
			p.next()
		}
	}
	return n
}

func (p *parser) simplestmt() Node {
	switch p.tok {
	case _Name:
		name := p.nameExpr()

		if p.tok != _Assign && p.tok != _Comma {
			p.unexpected(p.tok)
			p.advance(_Semi)
		}
		return p.assign(name)
	case _Literal:
		return p.literal()
	case _Ref:
		return p.ref()
	default:
		p.unexpected(p.tok)
		p.advance(_Semi)
	}
	return nil
}

func (p *parser) initExpr() Node {
	var n Node

	switch p.tok {
	case _Name:
		name := p.nameExpr()

		if p.tok != _Assign && p.tok != _Comma {
			p.unexpected(p.tok)
			p.advance(_Semi)
		}
		return p.assign(name)
	case _Literal:
		n = p.literal()
	case _Ref:
		n = p.ref()
	default:
		p.unexpected(p.tok)
		p.advance(_Semi)
		return n
	}

	if p.tok == _Op {
		n = p.binaryExpr(n, 0)
	}
	return n
}

func (p *parser) forstmt() *ForStmt {
	nodpos := p.node()

	if !p.got(_For) {
		return nil
	}

	n := &ForStmt{
		node: nodpos,
	}

	if p.tok != _Lbrace {
		n.Init = p.initExpr()

		if !p.got(_Semi) {
			if p.tok == _Lbrace {
				n.Cond = n.Init
				n.Init = nil
				goto body
			}

			p.err("expected for loop condition")
			return nil
		}

		n.Cond = p.expr()
		p.want(_Semi)
		n.Post = p.simplestmt()
	}

body:
	n.Body = p.blockstmt()
	return n
}

func (p *parser) chain(cmd *CommandStmt) *ChainExpr {
	n := &ChainExpr{
		Commands: []*CommandStmt{cmd},
	}

	for p.tok != _Semi && p.tok != _EOF {
		if p.tok != _Name {
			p.expected(_Name)
			p.advance(_Semi)
			continue
		}

		n.Commands = append(n.Commands, p.command(p.name()))

		if !p.got(_Arrow) && p.tok != _Semi && p.tok != _EOF {
			break
		}
	}
	return n
}

// operand parses an operand, this will either be a literal, variable reference,
// object, or an array.
func (p *parser) operand() Node {
	var n Node

	switch p.tok {
	case _Literal:
		n = p.literal()
	case _Ref:
		n = p.ref()
	case _Lparen:
		n = p.obj()
	case _Lbrack:
		n = p.arr()
	}
	return n
}

func (p *parser) expr() Node {
	return p.binaryExpr(nil, 0)
}

// binaryExpr parses a binary expression. First this will parse a normal
// expression then attempt to parse a binary expression if any operator
// tokens are encountered using the given operator precedence of prec.
func (p *parser) binaryExpr(n Node, prec int) Node {
	if n == nil {
		n = p.unaryExpr()
	}

	for p.tok == _Op && p.prec > prec {
		o := &Operation{
			node: p.node(),
			Op:   p.op,
		}

		oprec := p.prec

		p.next()

		o.Left = n
		o.Right = p.binaryExpr(nil, oprec)

		n = o
	}
	return n
}

// unaryExpr parses a unary expression, an expression with only a single
// operand.
func (p *parser) unaryExpr() Node {
	if p.tok == _Name {
		n := p.command(p.name())

		if p.got(_Arrow) {
			return p.chain(n)
		}
		return n
	}
	return p.operand()
}

func (p *parser) command(name *Name) *CommandStmt {
	n := &CommandStmt{
		node: name.node,
		Name: name,
	}

	for p.tok != _Arrow && p.tok != _Semi && p.tok != _EOF {
		if p.tok == _Name {
			n.Args = append(n.Args, p.name())
			continue
		}

		arg := p.operand()

		if arg == nil {
			break
		}
		n.Args = append(n.Args, arg)
	}
	return n
}

func (p *parser) assign(first Node) Node {
	n := &AssignStmt{
		node: node{pos: first.Pos()},
	}

	left := []Node{first}

	for p.got(_Comma) {
		left = append(left, p.nameExpr())
	}

	n.Left = &ExprList{
		node:  n.node,
		Nodes: left,
	}

	if !p.got(_Assign) {
		return nil
	}

	right := []Node{p.expr()}

	for p.got(_Comma) {
		right = append(right, p.expr())
	}

	n.Right = &ExprList{
		node:  node{pos: right[0].Pos()},
		Nodes: right,
	}
	return n
}

// stmt parses a top-level statement. If inRepl is true then this will allow
// for the parsing of variable reference expressions, as in a REPL you may want
// to have the contents of these displayed.
func (p *parser) stmt(inRepl bool) Node {
	var n Node

	switch p.tok {
	case _Name:
		expr := p.nameExpr()

		if p.tok == _Assign || p.tok == _Comma {
			n = p.assign(expr)
			break
		}

		var name *Name

		switch v := expr.(type) {
		case *IndExpr:
			p.errAt(name.Pos(), "unassigned index expression")
			p.advance(_Semi)
		case *Name:
			name = v
		}

		cmd := p.command(name)
		n = cmd

		if p.got(_Arrow) {
			n = p.chain(cmd)
		}
	case _Break, _Continue:
		n = &BranchStmt{
			node: p.node(),
			Tok:  p.tok,
		}
		p.next()
	case _Match:
		n = p.matchstmt()
		return n
	case _If:
		n = p.ifstmt()
		return n
	case _For:
		n = p.forstmt()
		return n
	case _Ref:
		if inRepl {
			n = p.ref()
			break
		}
		fallthrough
	default:
		p.unexpected(p.tok)
		p.advance(_Semi)
	}

	if !p.got(_Semi) {
		if !inRepl {
			p.expected(_Semi)
		}
	}
	return n
}

func (p *parser) parse(inRepl bool) ([]Node, error) {
	nn := make([]Node, 0)

	for p.tok != _EOF {
		nn = append(nn, p.stmt(inRepl))
	}

	if p.errc > 0 {
		return nil, fmt.Errorf("parser encountered %d error(s)", p.errc)
	}
	return nn, nil
}
