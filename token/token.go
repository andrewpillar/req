package token

import "strconv"

type Pos struct {
	File string
	Line int
	Col  int
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

	Open  // open
	Env   // env
	Exit  // exit
	Write // write

	HEAD    // HEAD
	OPTIONS // OPTIONS
	GET     // GET
	POST    // POST
	PUT     // PUT
	PATCH   // PATCH
	DELETE  // DELETE
)

type Type uint

//go:generate stringer -type Type -linecomment
const (
	String Type = iota + 1 // string
	Int                    // int
	Bool                   // bool
)

var keywords = map[string]Token{
	"if":      If,
	"else":    Else,
	"match":   Match,
	"range":   Range,
	"yield":   Yield,
	"open":    Open,
	"env":     Env,
	"exit":    Exit,
	"write":   Write,
	"HEAD":    HEAD,
	"OPTIONS": OPTIONS,
	"GET":     GET,
	"POST":    POST,
	"PUT":     PUT,
	"PATCH":   PATCH,
	"DELETE":  DELETE,
}

func Lookup(s string) Token {
	if tok, ok := keywords[s]; ok {
		return tok
	}
	return Name
}
