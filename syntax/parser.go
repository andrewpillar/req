package syntax

import (
	"fmt"
	"os"

	"github.com/andrewpillar/req/token"
)

type parser struct {
	*scanner

	errc int
}

func ParseFile(fname string, errh func(token.Pos, string)) ([]*Node, error) {
	f, err := os.Open(fname)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	p := parser{
		scanner: newScanner(newSource(fname, f, errh)),
	}

	if p.errc > 0 {
		return nil, fmt.Errorf("parser encountered %d error(s)", p.errc)
	}
	return p.parse(), nil
}

func (p *parser) advance(tok token.Token) {
	for p.tok != tok {
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

func (p *parser) want(tok, next token.Token) {
	if !p.got(tok) {
		p.expected(tok)
		p.advance(next)
	}
}

func (p *parser) node(op Op) *Node {
	return &Node{
		Op:  op,
		Pos: p.pos,
	}
}

func (p *parser) name() *Node {
	if p.tok != token.Name {
		return nil
	}

	n := p.node(ONAME)
	n.Value = p.lit
	p.next()
	return n
}

func (p *parser) literal() *Node {
	if p.tok != token.Literal {
		return nil
	}

	n := p.node(OLIT)
	n.Type = p.typ
	n.Value = p.lit
	p.next()
	return n
}

func (p *parser) ref() *Node {
	if p.tok != token.Ref {
		return nil
	}

	p.got(token.Ref)

	n := p.node(OREF)
	n.Left = p.name()

loop:
	for {
		pos := p.pos

		switch p.tok {
		case token.Dot:
			p.next()

			if p.tok != token.Name {
				p.expected(token.Name)
				p.advance(token.Semi)
				return nil
			}

			tmp := p.node(OREFDOT)
			tmp.Pos = pos
			tmp.Left = n
			tmp.Right = p.name()

			n = tmp
		case token.Lbrack:
			p.next()

			if p.tok == token.Rbrack {
				p.err("expected string, int, or variable")
				p.next()
				break
			}

			tmp := p.node(OREFIND)
			tmp.Pos = pos
			tmp.Left = n

			switch p.tok {
			case token.Literal:
				tmp.Right = p.literal()
			case token.Ref:
				tmp.Right = p.ref()
			default:
				p.unexpected(p.tok)
				p.next()
			}

			n = tmp
		default:
			break loop
		}
	}
	return n
}

func (p *parser) obj() *Node {
	p.want(token.Lbrace, token.Semi)

	n := p.node(OOBJ)

	for p.tok != token.Rbrace {
		if p.tok != token.Name {
			p.expected(token.Name)
			p.advance(token.Comma)
			p.got(token.Comma)
			continue
		}

		key := p.node(OKEY)
		key.Left = p.name()

		if p.tok != token.Colon {
			p.expected(token.Colon)
			p.advance(token.Comma)
			goto cont
		}

		p.got(token.Colon)

		key.Right = p.operand()

		n.InsertBody(key)

cont:
		if p.tok != token.Comma && p.tok != token.Rbrace {
			p.err("expected comma or }")
			p.advance(token.Semi)
			break
		}
		p.got(token.Comma)
	}

	p.want(token.Rbrace, token.Semi)
	return n
}

func (p *parser) arr() *Node {
	p.want(token.Lbrack, token.Semi)

	n := p.node(OARR)

	for p.tok != token.Rbrack {
		if p.tok != token.Literal {
			p.expected(token.Literal)
			p.advance(token.Comma)
			goto cont
		}

		n.InsertList(p.literal())

cont:
		if p.tok != token.Comma && p.tok != token.Rbrack {
			p.err("expected comma or ]")
			p.advance(token.Semi)
			break
		}
		p.got(token.Comma)
	}

	p.want(token.Rbrack, token.Semi)
	return n
}

func (p *parser) operand() *Node {
	var n *Node

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
	}

	return n
}

func (p *parser) blockstmt() *Node {
	p.want(token.Lbrace, token.Semi)

	n := p.node(OBLOCK)

	for p.tok != token.Rbrace {
		n.InsertBody(p.stmt())
	}

	p.want(token.Rbrace, token.Semi)
	return n
}

func (p *parser) yield() *Node {
	if !p.got(token.Yield) {
		return nil
	}

	n := p.node(OYIELD)
	n.Left = p.operand()

	return n
}

func (p *parser) exit() *Node {
	if !p.got(token.Exit) {
		return nil
	}

	n := p.node(OEXIT)
	n.Left = p.operand()

	return n
}

func (p *parser) casestmt() *Node {
	n := p.node(OCASE)

	if p.tok == token.Name {
		if p.lit != "_" {
			p.unexpected(p.tok)
			return nil
		}

		n.Left = p.name()
		goto right
	}

	n.Left = p.literal()

right:
	p.want(token.Arrow, token.Comma)

	switch p.tok {
	case token.Lbrace:
		n.Right = p.blockstmt()
	case token.Yield:
		n.Right = p.yield()
	}
	return n
}

func (p *parser) matchstmt() *Node {
	if p.tok != token.Match {
		return  nil
	}

	n := p.node(OMATCH)

	p.next()

	switch p.tok {
	case token.Literal:
		n.Left = p.literal()
	case token.Ref:
		n.Left = p.ref()
	default:
		p.unexpected(p.tok)
		return nil
	}

	p.want(token.Lbrace, token.Rbrace)

	for p.tok != token.Rbrace {
		n.InsertBody(p.casestmt())

		if p.tok != token.Comma && p.tok != token.Rbrace {
			p.err("expected comma or }")
			break
		}
		p.got(token.Comma)
	}

	p.got(token.Rbrace)
	return n
}

func (p *parser) expr() *Node {
	switch p.tok {
	case token.Match:
		return p.matchstmt()
	default:
		return p.operand()
	}
}

func (p *parser) action(op Op, val string, hasArrow bool) *Node {
	n := p.node(op)
	n.Value = val

	p.next()

	end := token.Semi

	if hasArrow {
		end = token.Arrow
	}

	n.Left = p.node(OLIST)

	for p.tok != end {
		n.Left.InsertList(p.operand())
	}

	if hasArrow {
		p.want(token.Arrow, token.Semi)
		n.Right = p.expr()
	}
	return n
}

func (p *parser) open() *Node {
	if p.tok != token.Open {
		return nil
	}
	return p.action(OOPEN, "", false)
}

func (p *parser) env() *Node {
	if p.tok != token.Env {
		return nil
	}
	return p.action(OENV, "", false)
}

func (p *parser) method() * Node {
	toks := map[token.Token]struct{}{
		token.HEAD:    {},
		token.OPTIONS: {},
		token.GET:     {},
		token.POST:    {},
		token.PUT:     {},
		token.PATCH:   {},
		token.DELETE:  {},
	}

	if _, ok := toks[p.tok]; !ok {
		return nil
	}
	return p.action(OMETHOD, p.lit, true)
}

func (p *parser) vardecl() *Node {
	if p.tok != token.Name {
		return nil
	}

	n := p.node(OVAR)
	n.Left = p.name()

	if !p.got(token.Assign) {
		return nil
	}

	switch p.tok {
	case token.Open:
		n.Right = p.open()
	case token.Env:
		n.Right = p.env()
	case token.HEAD, token.OPTIONS, token.GET, token.POST, token.PUT, token.PATCH, token.DELETE:
		n.Right = p.method()
	default:
		n.Right = p.expr()
	}

	return n
}

func (p *parser) write() *Node {
	if p.tok != token.Write {
		return nil
	}
	return p.action(OWRITE, "", true)
}

func (p *parser) stmt() *Node {
	var n *Node

	switch p.tok {
	case token.Semi:
		p.next()
		break
	case token.Name:
		n = p.vardecl()
	case token.Write:
		n = p.write()
	case token.Yield:
		n = p.yield()
	case token.Exit:
		n = p.exit()
	default:
		p.unexpected(p.tok)
		p.next()
	}

	return n
}

func (p *parser) parse() []*Node {
	nn := make([]*Node, 0)

	for p.tok != token.EOF {
		nn = append(nn, p.stmt())
	}
	return nn
}
