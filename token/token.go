package token

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
		s += "," + strconv.FormatInt(int64(p.Line), 10)

		if p.Col > 0 {
			s += ":" + strconv.FormatInt(int64(p.Col), 10)
		}
	}
	return s
}

type Token uint

//go:generate stringer -type Token -linecomment
const (
	EOF Token = iota + 1 // eof

	Name    // name
	Literal // literal

	Semi   // semi or newline
	Comma  // ,
	Colon  // :
	Dot    // .
	DotDot // ..
	Arrow  // ->
	Assign // =
	Ref    // $

	Lbrace // {
	Rbrace // }
	Lbrack // [
	Rbrack // ]

	If    // if
	Else  // else
	Match // match
	Range // range
	Yield // yield
)

type Type uint

//go:generate stringer -type Type -linecomment
const (
	String Type = iota + 1 // string
	Int                    // int
	Bool                   // bool
)

var keywords = map[string]Token{
	"if":    If,
	"else":  Else,
	"match": Match,
	"range": Range,
	"yield": Yield,
}

func Lookup(s string) Token {
	if tok, ok := keywords[s]; ok {
		return tok
	}
	return Name
}
