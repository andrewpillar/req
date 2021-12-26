package syntax

import (
	"errors"
	"io"
	"unicode/utf8"

	"github.com/andrewpillar/req/token"
)

type source struct {
	name        string
	r           io.Reader
	pos0, pos   int
	eof         int
	line0, line int
	col0, col   int
	errh        func(token.Pos, string)
	buf         []byte
	lit         int
}

func newSource(name string, r io.Reader, errh func(token.Pos, string)) *source {
	return &source{
		name: name,
		r:    r,
		line: 1,
		errh: errh,
		lit:  -1,
		buf:  make([]byte, 4096),
	}
}

func (s *source) getpos() token.Pos {
	return token.Pos{
		File: s.name,
		Line: s.line,
		Col:  s.col,
	}
}

func (s *source) err(msg string) {
	s.errh(s.getpos(), msg)
}

func (s *source) get() rune {
redo:
	s.pos0, s.line0, s.col0 = s.pos, s.line, s.col

	if s.pos == 0 || s.pos >= len(s.buf) {
		if s.lit >= 0 {
			buf := s.buf[s.lit:s.pos]

			s.buf = make([]byte, len(s.buf)+len(buf))
			copy(s.buf, buf)
		}

		n, err := s.r.Read(s.buf)

		if err != nil {
			if !errors.Is(err, io.EOF) {
				s.err("io error: " + err.Error())
			}
			return -1
		}

		s.pos = 0
		s.eof = n
	}

	if s.pos == s.eof {
		return -1
	}

	b := s.buf[s.pos]

	if b >= utf8.RuneSelf {
		r, w := utf8.DecodeRune(s.buf[s.pos:])

		s.pos += w
		s.col += w

		return r
	}

	s.pos++
	s.col++

	if b == 0 {
		s.err("invalid NUL byte")
		goto redo
	}

	if b == '\n' {
		s.line++
		s.col = 0
	}
	return rune(b)
}

func (s *source) unget() {
	s.pos, s.line, s.col = s.pos0, s.line0, s.col0
}

func (s *source) startLit() {
	s.lit = s.pos0
}

func (s *source) stopLit() string {
	if s.lit < 0 {
		panic("syntax: negative literal position")
	}

	lit := s.buf[s.lit:s.pos]
	s.lit = -1

	return string(lit)
}
