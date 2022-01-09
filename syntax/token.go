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

func (p Pos) Err(msg string) error {
	return errors.New(p.String() + " - " + msg)
}

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
	_DotDot // ..
	_Arrow  // ->
	_Assign // =
	_Ref    // $

	_Lbrace // {
	_Rbrace // }
	_Lbrack // [
	_Rbrack // ]

	_If    // if
	_Else  // else
	_Match // match
	_Range // range
)

type LitType uint

//go:generate stringer -type LitType -linecomment
const (
	StringLit LitType = iota + 1 // string
	IntLit                       // int
	BoolLit                      // bool
)

var keywords = map[string]token{
	"if":    _If,
	"else":  _Else,
	"match": _Match,
	"range": _Range,
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
