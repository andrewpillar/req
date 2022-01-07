package syntax

import "fmt"

type scanner struct {
	*source

	pos Pos
	op  Op
	tok token
	typ LitType
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

	sc.tok = lookupTok(lit)
	sc.lit = lit
}

func (sc *scanner) number() {
	sc.startLit()

	r := sc.get()

	for isDigit(r) {
		r = sc.get()
	}
	sc.unget()

	sc.tok = _Literal
	sc.typ = IntLit
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

	sc.tok = _Literal
	sc.typ = StringLit
	sc.lit = lit[1 : len(lit)-1]
}

func (sc *scanner) next() {
redo:
	sc.op = Op(0)
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

		if ok := isop(sc.lit); ok {
			sc.tok = _Op
			sc.lit = ""
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
		if sc.get() == '.' {
			sc.tok = _DotDot
			break
		}
		sc.unget()
		sc.tok = _Dot
	case '{':
		sc.tok = _Lbrace
	case '}':
		sc.tok = _Rbrace
	case '[':
		sc.tok = _Lbrack
	case ']':
		sc.tok = _Rbrack
	case '=':
		if sc.get() == '=' {
			sc.tok = _Op
			sc.op = EqOp
			break
		}
		sc.unget()
		sc.tok = _Assign
	case '!':
		if sc.get() == '=' {
			sc.tok = _Op
			sc.op = NeqOp
			break
		}
		sc.unget()
	case '<':
		if sc.get() == '=' {
			sc.tok = _Op
			sc.op = LeqOp
			break
		}
		sc.unget()
		sc.op = LtOp
	case '>':
		if sc.get() == '=' {
			sc.tok = _Op
			sc.op = GeqOp
			break
		}
		sc.unget()
		sc.op = GtOp
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
