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

func Parse(name string, r io.Reader, errh func(Pos, string)) ([]Node, error) {
	p := parser{
		scanner: newScanner(newSource(name, r, errh)),
	}
	return p.parse()
}

func ParseFile(fname string, errh func(Pos, string)) ([]Node, error) {
	f, err := os.Open(fname)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	return Parse(fname, f, errh)
}

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

func (p *parser) got(tok token) bool {
	if p.tok == tok {
		p.next()
		return true
	}
	return false
}

func (p *parser) errAt(pos Pos, msg string) {
	p.errc++
	p.scanner.source.errh(pos, msg)
}

func (p *parser) err(msg string) {
	p.errAt(p.pos, msg)
}

func (p *parser) expected(tok token) {
	p.err("expected " + tok.String())
}

func (p *parser) unexpected(tok token) {
	p.err("unexpected " + tok.String())
}

func (p *parser) want(tok token) {
	if !p.got(tok) {
		p.expected(tok)
	}
}

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
	p.want(_Lbrace)

	n := &Object{
		node: p.node(),
	}

	p.list(_Comma, _Rbrace, func() {
		if p.tok != _Name {
			p.expected(_Name)
			p.advance(_Rbrace, _Semi)
			return
		}

		key := p.name()

		p.want(_Colon)

		val := p.operand()

		n.Pairs = append(n.Pairs, &KeyExpr{
			node:  p.node(),
			Key:   key,
			Value: val,
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
		if p.tok != _Literal {
			p.expected(_Literal)
			p.next()
			return
		}
		n.Items = append(n.Items, p.literal())
	})
	return n
}

func (p *parser) blockstmt() *BlockStmt {
	p.want(_Lbrace)

	n := &BlockStmt{
		node: p.node(),
	}

	for p.tok != _Rbrace && p.tok != _EOF {
		n.Nodes = append(n.Nodes, p.stmt(true))
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

func (p *parser) infixexpr() Node {
	n := p.expr()

	for p.tok == _Op {
		o := &Operation{
			node: node{pos: n.Pos()},
			Op:   p.op,
			Left: n,
		}
		p.next()
		o.Right = p.infixexpr()

		n = o
	}
	return n
}

func (p *parser) ifstmt() *IfStmt {
	if !p.got(_If) {
		return nil
	}

	n := &IfStmt{
		node: p.node(),
		Cond: p.infixexpr(),
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
			p.err("expected " + _Arrow.String() + " or " + _Semi.String())
			p.next()
		}
	}
	return n
}

func (p *parser) operand() Node {
	var n Node

	switch p.tok {
	case _Literal:
		n = p.literal()
	case _Ref:
		n = p.ref()
	case _Lbrace:
		n = p.obj()
	case _Lbrack:
		n = p.arr()
	default:
		p.unexpected(p.tok)
		p.advance(_Rbrace, _Semi)
	}
	return n
}

func (p *parser) expr() Node {
	switch p.tok {
	case _Name:
		n := p.command(p.name())

		if p.got(_Arrow) {
			return p.chain(n)
		}
		return n
	default:
		return p.operand()
	}
}

func (p *parser) command(name *Name) *CommandStmt {
	n := &CommandStmt{
		node: name.node,
		Name: name,
	}

	for p.tok != _Arrow && p.tok != _Semi && p.tok != _EOF {
		n.Args = append(n.Args, p.operand())
	}
	return n
}

func (p *parser) vardecl(name *Name) *VarDecl {
	n := &VarDecl{
		node: name.node,
		Name: name,
	}

	if !p.got(_Assign) {
		return nil
	}

	n.Value = p.expr()

	return n
}

func (p *parser) stmt(inBlock bool) Node {
	var n Node

	switch p.tok {
	case _Name:
		name := p.name()

		if p.tok == _Assign {
			n = p.vardecl(name)
			break
		}

		cmd := p.command(name)
		n = cmd

		if p.got(_Arrow) {
			n = p.chain(cmd)
		}
	case _Match:
		n = p.matchstmt()
		return n
	case _If:
		n = p.ifstmt()
		return n
	default:
		p.unexpected(p.tok)
		p.advance(_Semi)
	}

	if p.tok != _EOF {
		if !p.got(_Semi) {
			p.expected(_Semi)
		}
	}
	return n
}

func (p *parser) parse() ([]Node, error) {
	nn := make([]Node, 0)

	for p.tok != _EOF {
		nn = append(nn, p.stmt(false))
	}

	if p.errc > 0 {
		return nil, fmt.Errorf("parser encountered %d error(s)", p.errc)
	}
	return nn, nil
}
