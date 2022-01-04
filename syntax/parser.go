package syntax

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/andrewpillar/req/token"
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
		scanner: newScanner(newSource("", strings.NewReader(s), func(pos token.Pos, msg string) {
			errs = append(errs, errors.New(msg))
		})),
	}

	if p.tok != token.Ref {
		return nil, errors.New("expected $")
	}

	n := p.ref()

	if len(errs) > 0 {
		return nil, errs[0]
	}
	return n, nil
}

func ParseFile(fname string, errh func(token.Pos, string)) ([]Node, error) {
	f, err := os.Open(fname)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	p := parser{
		scanner: newScanner(newSource(fname, f, errh)),
	}
	return p.parse()
}

func (p *parser) advance(follow ...token.Token) {
	set := make(map[token.Token]struct{})

	for _, tok := range follow {
		set[tok] = struct{}{}
	}
	set[token.EOF] = struct{}{}

	for {
		if _, ok := set[p.tok]; ok {
			break
		}
		p.next()
	}
}

func (p *parser) got(tok token.Token) bool {
	if p.tok == tok {
		p.next()
		return true
	}
	return false
}

func (p *parser) errAt(pos token.Pos, msg string) {
	p.errc++
	p.scanner.source.errh(pos, msg)
}

func (p *parser) err(msg string) {
	p.errAt(p.pos, msg)
}

func (p *parser) expected(tok token.Token) {
	p.err("expected " + tok.String())
}

func (p *parser) unexpected(tok token.Token) {
	p.err("unexpected " + tok.String())
}

func (p *parser) want(tok token.Token) {
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
	if p.tok != token.Name {
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
	if p.tok != token.Literal {
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
	if p.tok != token.Ref {
		return nil
	}

	p.got(token.Ref)

	ref := &Ref{
		node: p.node(),
		Left: p.name(),
	}

loop:
	for {
		pos := p.pos

		switch p.tok {
		case token.Dot:
			p.next()

			if p.tok != token.Name {
				p.expected(token.Name)
				p.next()
				return nil
			}

			left := ref.Left

			ref.Left = &DotExpr{
				node:  node{pos: pos},
				Left:  left,
				Right: p.name(),
			}
		case token.Lbrack:
			p.next()

			if p.tok == token.Rbrack {
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
			case token.Literal:
				ind.Right = p.literal()
			case token.Ref:
				ind.Right = p.ref()
			default:
				p.unexpected(p.tok)
				p.next()
			}

			p.want(token.Rbrack)
			ref.Left = ind
		default:
			break loop
		}
	}
	return ref
}

func (p *parser) list(sep, end token.Token, parse func()) {
	for p.tok != token.EOF && p.tok != end {
		parse()

		if !p.got(sep) && p.tok != end {
			p.err("expected " + sep.String() + " or " + end.String())
			p.next()
		}
	}
	p.want(end)
}

func (p *parser) obj() *Object {
	p.want(token.Lbrace)

	n := &Object{
		node: p.node(),
	}

	p.list(token.Comma, token.Rbrace, func() {
		if p.tok != token.Name {
			p.expected(token.Name)
			p.advance(token.Rbrace, token.Semi)
			return
		}

		key := p.name()

		p.want(token.Colon)

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
	p.want(token.Lbrack)

	n := &Array{
		node: p.node(),
	}

	p.list(token.Comma, token.Rbrack, func() {
		if p.tok != token.Literal {
			p.expected(token.Literal)
			p.next()
			return
		}
		n.Items = append(n.Items, p.literal())
	})
	return n
}

func (p *parser) blockstmt() *BlockStmt {
	p.want(token.Lbrace)

	n := &BlockStmt{
		node: p.node(),
	}

	for p.tok != token.Rbrace && p.tok != token.EOF {
		n.Nodes = append(n.Nodes, p.stmt(true))
	}

	p.want(token.Rbrace)
	return n
}

func (p *parser) casestmt() *CaseStmt {
	n := &CaseStmt{
		node: p.node(),
	}

	if p.tok != token.Literal {
		p.unexpected(p.tok)
		p.next()
		return nil
	}

	n.Value = p.literal()

	p.want(token.Arrow)

	switch p.tok {
	case token.Lbrace:
		n.Then = p.blockstmt()
	case token.Name:
		n.Then = p.command(p.name())
	default:
		p.unexpected(p.tok)
		p.next()
	}
	return n
}

func (p *parser) matchstmt() *MatchStmt {
	if p.tok != token.Match {
		return nil
	}

	n := &MatchStmt{
		node: p.node(),
	}

	p.next()

	switch p.tok {
	case token.Literal:
		n.Cond = p.literal()
	case token.Ref:
		n.Cond = p.ref()
	default:
		p.unexpected(p.tok)
		p.next()
	}

	p.want(token.Lbrace)

	for p.tok != token.Rbrace {
		if p.tok == token.Name {
			if p.lit != "_" {
				p.unexpected(token.Name)
				p.advance(token.Rbrace, token.Semi)
				continue
			}

			p.name()
			p.want(token.Arrow)

			switch p.tok {
			case token.Lbrace:
				n.Default = p.blockstmt()
			case token.Name:
				n.Default = p.command(p.name())
			default:
				p.unexpected(p.tok)
				p.next()
			}
			p.got(token.Semi)
			continue
		}

		n.Cases = append(n.Cases, p.casestmt())

		p.got(token.Semi)
	}

	p.got(token.Rbrace)
	return n
}

func (p *parser) chain(cmd *CommandStmt) *ChainExpr {
	n := &ChainExpr{
		Commands: []*CommandStmt{cmd},
	}

	for p.tok != token.Semi && p.tok != token.EOF {
		if p.tok != token.Name {
			p.expected(token.Name)
			p.advance(token.Semi)
			continue
		}

		n.Commands = append(n.Commands, p.command(p.name()))

		if !p.got(token.Arrow) && p.tok != token.Semi && p.tok != token.EOF {
			p.err("expected " + token.Arrow.String() + " or " + token.Semi.String())
			p.next()
		}
	}
	return n
}

func (p *parser) operand() Node {
	var n Node

	switch p.tok {
	case token.Literal:
		n = p.literal()
	case token.Ref:
		n = p.ref()
	case token.Lbrace:
		n = p.obj()
	case token.Lbrack:
		n = p.arr()
	default:
		p.unexpected(p.tok)
		p.advance(token.Rbrace, token.Semi)
	}
	return n
}

func (p *parser) expr() Node {
	switch p.tok {
	case token.Name:
		n := p.command(p.name())

		if p.got(token.Arrow) {
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

	for p.tok != token.Arrow && p.tok != token.Semi && p.tok != token.EOF {
		n.Args = append(n.Args, p.operand())
	}
	return n
}

func (p *parser) vardecl(name *Name) *VarDecl {
	n := &VarDecl{
		node: p.node(),
		Name: name,
	}

	if !p.got(token.Assign) {
		return nil
	}

	n.Value = p.expr()

	return n
}

func (p *parser) stmt(inBlock bool) Node {
	var n Node

	switch p.tok {
	case token.Name:
		name := p.name()

		if p.tok == token.Assign {
			n = p.vardecl(name)
			break
		}

		cmd := p.command(name)
		n = cmd

		if p.got(token.Arrow) {
			n = p.chain(cmd)
		}
	case token.Match:
		n = p.matchstmt()
	case token.Ref:
		n = p.ref()
	default:
		p.unexpected(p.tok)
		p.advance(token.Semi)
	}

	if p.tok != token.EOF {
		if !p.got(token.Semi) {
			p.expected(token.Semi)
		}
	}
	return n
}

func (p *parser) parse() ([]Node, error) {
	nn := make([]Node, 0)

	for p.tok != token.EOF {
		nn = append(nn, p.stmt(false))
	}

	if p.errc > 0 {
		return nil, fmt.Errorf("parser encountered %d error(s)", p.errc)
	}
	return nn, nil
}
