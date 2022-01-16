package syntax

import (
	"errors"
	"io"
	"unicode/utf8"
)

// source represents a source file being parsed for tokens. This records the
// current and previous position in the buffer using pos and pos0 respectively
// as well ass line0, line and col0, col for for explicit positional information
// for error reporting.
//
// eof denotes where in the buffer the EOF occurs.
//
// lit denotes the start position of a literal that we want to copy from the
// underlying buffer. If lit is < 0 when a copy of a literal is made then the
// a panic will happen.
//
// The errh callback is called to handle the reporting of errors that may occur
// during parsing of a file.
type source struct {
	name        string
	r           io.Reader
	pos0, pos   int
	eof         int
	line0, line int
	col0, col   int
	errh        func(Pos, string)
	buf         []byte
	lit         int
}

// newSource returns a new source for the given reader. The name of the source
// should be used to uniquely identify it, for example a filename.
func newSource(name string, r io.Reader, errh func(Pos, string)) *source {
	return &source{
		name: name,
		r:    r,
		line: 1,
		errh: errh,
		lit:  -1,
		buf:  make([]byte, 4096),
	}
}

// getpos returns the current position in the source.
func (s *source) getpos() Pos {
	return Pos{
		File: s.name,
		Line: s.line,
		Col:  s.col,
	}
}

func (s *source) err(msg string) {
	s.errh(s.getpos(), msg)
}

// get returns the next rune in the source. If EOF has been reached then -1
// is returned. If a fatal error occurs when reading from the underlying
// source, then an error is recorded via errh and -1 is returned.
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

// unget moves the position of the source back one. This cannot be called
// multiple subsequent times.
func (s *source) unget() {
	s.pos, s.line, s.col = s.pos0, s.line0, s.col0
}

// startLit sets the literal position to pos0, the previous position in the
// buffer.
func (s *source) startLit() {
	s.lit = s.pos0
}

// stopLit returns the literal being scanned from the buffer. If the literal
// position is < 0 then this panics with "syntax: negative literal position".
func (s *source) stopLit() string {
	if s.lit < 0 {
		panic("syntax: negative literal position")
	}

	lit := s.buf[s.lit:s.pos]
	s.lit = -1

	return string(lit)
}
