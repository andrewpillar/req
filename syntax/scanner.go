package syntax

import (
	"fmt"
	"unicode"
)

type scanner struct {
	*source

	pos  Pos
	op   Op
	prec int
	tok  token
	typ  LitType
	lit  string
}

func newScanner(src *source) *scanner {
	sc := &scanner{
		source: src,
	}
	sc.next()
	return sc
}

func isLetter(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || r == '_' || unicode.IsLetter(r)
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

	sc.tok = lookupTok(lit)
	sc.lit = lit
}

func (sc *scanner) number() {
	sc.startLit()

	r := sc.get()

	isFloat := false
	typ := IntLit

	for {
		if !isDigit(r) {
			if r == '.' {
				if isFloat {
					sc.err("invalid point in float")
					break
				}

				isFloat = true
				r = sc.get()
				continue
			}
			break
		}
		r = sc.get()
	}
	sc.unget()

	if isFloat {
		typ = FloatLit
	}

	sc.tok = _Literal
	sc.typ = typ
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

		if r == '$' {
			if sc.get() == '(' {
				interpolate = true
			}
			sc.unget()
		}

		if r == ')' {
			if interpolate {
				interpolate = false
			}
		}
		r = sc.get()
	}

	lit := sc.stopLit()

	sc.tok = _Literal
	sc.typ = StringLit
	sc.lit = lit[1 : len(lit)-1]
}

func (sc *scanner) next() {
redo:
	sc.op = Op(0)
	sc.prec = 0
	sc.tok = token(0)
	sc.lit = sc.lit[0:0]
	sc.typ = LitType(0)

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

		if op, ok := lookupOp(sc.lit); ok {
			sc.op = op
			sc.prec = opwordprec[op]
			sc.tok = _Op
			sc.lit = ""
		}

		if sc.lit == "true" || sc.lit == "false" {
			sc.tok = _Literal
			sc.typ = BoolLit
		}
		return
	}

	if isDigit(r) {
		sc.number()
		return
	}

	switch r {
	case -1:
		sc.tok = _EOF
	case ';', '\n':
		sc.tok = _Semi
	case ',':
		sc.tok = _Comma
	case ':':
		sc.tok = _Colon
	case '.':
		sc.tok = _Dot
	case '{':
		sc.tok = _Lbrace
	case '}':
		sc.tok = _Rbrace
	case '(':
		sc.tok = _Lparen
	case ')':
		sc.tok = _Rparen
	case '[':
		sc.tok = _Lbrack
	case ']':
		sc.tok = _Rbrack
	case '=':
		if sc.get() == '=' {
			sc.tok = _Op
			sc.op, sc.prec = EqOp, precCmp
			break
		}
		sc.unget()
		sc.tok = _Assign
	case '!':
		if sc.get() == '=' {
			sc.tok = _Op
			sc.op, sc.prec = NeqOp, precCmp
			break
		}
		sc.unget()
	case '<':
		sc.tok = _Op

		if sc.get() == '=' {
			sc.op, sc.prec = LeqOp, precCmp
			break
		}
		sc.unget()
		sc.op, sc.prec = LtOp, precCmp
	case '>':
		if sc.get() == '=' {
			sc.tok = _Op
			sc.op, sc.prec = GeqOp, precCmp
			break
		}
		sc.unget()
		sc.tok = _Op
		sc.op, sc.prec = GtOp, precCmp
	case '$':
		sc.tok = _Ref
	case '"':
		sc.string()
	case '-':
		if sc.get() == '>' {
			sc.tok = _Arrow
			break
		}
		sc.unget()
	default:
		sc.err(fmt.Sprintf("unexpected token %U", r))
	}
}
