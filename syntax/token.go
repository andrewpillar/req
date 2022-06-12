package syntax

import (
	"errors"
	"strconv"
)

type Pos struct {
	File string
	Line int
	Col  int
}

// Err records an error at the given position.
func (p Pos) Err(msg string) error {
	return errors.New(p.String() + " - " + msg)
}

// String formats the position into a string. The string will be formatted like
// so depending on the values present,
//
// File
// File,Line
// File,Line:Col
func (p Pos) String() string {
	s := p.File

	if p.Line > 0 {
		if p.File != "" {
			s += ","
		}

		s += strconv.FormatInt(int64(p.Line), 10)

		if p.Col > 0 {
			s += ":" + strconv.FormatInt(int64(p.Col), 10)
		}
	}
	return s
}

type Op uint

//go:generate stringer -type Op -linecomment
const (
	EqOp  Op = iota + 1 // ==
	NeqOp               // !=
	LtOp                // <
	LeqOp               // <=
	GtOp                // >
	GeqOp               // >=

	// pseudo-operators
	InOp  // in
	AndOp // and
	OrOp  // or
)

const (
	precOr = iota + 1
	precAnd
	precCmp
	precIn
)

type token uint

//go:generate stringer -type token -linecomment
const (
	_EOF token = iota + 1 // eof

	_Name    // name
	_Literal // literal

	_Op // op

	_Semi   // semi or newline
	_Comma  // ,
	_Colon  // :
	_Dot    // .
	_Arrow  // ->
	_Assign // =
	_Ref    // $

	_Lbrace // {
	_Rbrace // }
	_Lparen // )
	_Rparen // )
	_Lbrack // [
	_Rbrack // ]

	_Break    // break
	_Continue // continue
	_If       // if
	_Else     // else
	_For      // for
	_Match    // match
	_Range    // range
)

type LitType uint

//go:generate stringer -type LitType -linecomment
const (
	StringLit LitType = iota + 1 // string
	IntLit                       // int
	FloatLit                     // float
	DurationLit                  // duration
	BoolLit                      // bool
)

var keywords = map[string]token{
	"break":    _Break,
	"continue": _Continue,
	"if":       _If,
	"else":     _Else,
	"for":      _For,
	"match":    _Match,
	"range":    _Range,
}

func lookupTok(s string) token {
	if tok, ok := keywords[s]; ok {
		return tok
	}
	return _Name
}

var (
	opwords = map[string]Op{
		"in":  InOp,
		"and": AndOp,
		"or":  OrOp,
	}

	opwordprec = map[Op]int{
		InOp:  precIn,
		AndOp: precAnd,
		OrOp:  precOr,
	}
)

func lookupOp(s string) (Op, bool) {
	if op, ok := opwords[s]; ok {
		return op, ok
	}
	return Op(0), false
}
