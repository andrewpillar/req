package syntax

import (
	"fmt"

	"github.com/andrewpillar/req/token"
)

type scanner struct {
	*source

	pos token.Pos
	tok token.Token
	typ token.Type
	lit string
}

func newScanner(src *source) *scanner {
	sc := &scanner{
		source: src,
	}
	sc.next()
	return sc
}

func isLetter(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || r == '_'
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func (sc *scanner) ident() {
	sc.startLit()

	r := sc.get()

	for isLetter(r) || isDigit(r) || r == '-' {
		r = sc.get()
	}
	sc.unget()

	lit := sc.stopLit()

	sc.tok = token.Lookup(lit)
	sc.lit = lit

	if sc.lit == "true" || sc.lit == "false" {
		sc.tok = token.Literal
		sc.typ = token.Bool
	}
}

func (sc *scanner) number() {
	sc.startLit()

	r := sc.get()

	for isDigit(r) {
		r = sc.get()
	}
	sc.unget()

	sc.tok = token.Literal
	sc.typ = token.Int
	sc.lit = sc.stopLit()
}

func (sc *scanner) string() {
	sc.startLit()

	interpolate := false
	r := sc.get()

	for {
		if r == '"' {
			if !interpolate {
				break
			}
		}
		if r == '\\' {
			r = sc.get()

			if r == '"' {
				r = sc.get()
			}
			continue
		}
		if r == '\n' {
			sc.err("unexpected newline in string")
			break
		}
		if r == '{' {
			interpolate = true
		}
		if r == '}' {
			interpolate = false
		}
		r = sc.get()
	}

	lit := sc.stopLit()

	sc.tok = token.Literal
	sc.typ = token.String
	sc.lit = lit[1 : len(lit)-1]
}

func (sc *scanner) next() {
redo:
	sc.lit = sc.lit[0:0]
	sc.typ = token.Type(0)

	r := sc.get()

	for r == ' ' || r == '\t' || r == '\r' || r == '\n' {
		r = sc.get()
	}

	if r == '#' {
		for r != '\n' {
			r = sc.get()
		}
		goto redo
	}

	sc.pos = sc.source.getpos()

	if isLetter(r) {
		sc.ident()
		return
	}

	if isDigit(r) {
		sc.number()
		return
	}

	switch r {
	case -1:
		sc.tok = token.EOF
	case ';', '\n':
		sc.tok = token.Semi
	case ',':
		sc.tok = token.Comma
	case ':':
		sc.tok = token.Colon
	case '.':
		if sc.get() == '.' {
			sc.tok = token.DotDot
			break
		}
		sc.unget()
		sc.tok = token.Dot
	case '{':
		sc.tok = token.Lbrace
	case '}':
		sc.tok = token.Rbrace
	case '[':
		sc.tok = token.Lbrack
	case ']':
		sc.tok = token.Rbrack
	case '=':
		sc.tok = token.Assign
	case '$':
		sc.tok = token.Ref
	case '"':
		sc.string()
	case '-':
		if sc.get() == '>' {
			sc.tok = token.Arrow
			break
		}
		sc.unget()
	default:
		sc.err(fmt.Sprintf("unexpected token %U", r))
	}
}
