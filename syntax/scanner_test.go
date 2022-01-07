package syntax

import (
	"os"
	"path/filepath"
	"testing"
)

func errh(t *testing.T) func(pos Pos, msg string) {
	return func(pos Pos, msg string) {
		t.Errorf("%s - %s\n", pos, msg)
	}
}

func Test_Scanner(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "gh.req"))

	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	sc := newScanner(newSource(f.Name(), f, errh(t)))

	tests := []struct {
		op  Op
		tok token
		typ LitType
		lit string
	}{
		{Op(0), _Name, LitType(0), "Stdout"},
		{Op(0), _Assign, LitType(0), ""},
		{Op(0), _Name, LitType(0), "open"},
		{Op(0), _Literal, StringLit, "/dev/stdout"},
		{Op(0), _Semi, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Stderr"},
		{Op(0), _Assign, LitType(0), ""},
		{Op(0), _Name, LitType(0), "open"},
		{Op(0), _Literal, StringLit, "/dev/stderr"},
		{Op(0), _Semi, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Endpoint"},
		{Op(0), _Assign, LitType(0), ""},
		{Op(0), _Literal, StringLit, "https://api.github.com"},
		{Op(0), _Semi, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Token"},
		{Op(0), _Assign, LitType(0), ""},
		{Op(0), _Name, LitType(0), "env"},
		{Op(0), _Literal, StringLit, "GH_TOKEN"},
		{Op(0), _Semi, LitType(0), ""},
		{Op(0), _If, LitType(0), "if"},
		{Op(0), _Ref, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Token"},
		{EqOp, _Op, LitType(0), ""},
		{Op(0), _Literal, StringLit, ""},
		{Op(0), _Lbrace, LitType(0), ""},
		{Op(0), _Name, LitType(0), "print"},
		{Op(0), _Literal, StringLit, "GH_TOKEN not set"},
		{Op(0), _Ref, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Stderr"},
		{Op(0), _Semi, LitType(0), ""},
		{Op(0), _Name, LitType(0), "exit"},
		{Op(0), _Literal, IntLit, "1"},
		{Op(0), _Semi, LitType(0), ""},
		{Op(0), _Rbrace, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Resp"},
		{Op(0), _Assign, LitType(0), ""},
		{Op(0), _Name, LitType(0), "GET"},
		{Op(0), _Literal, StringLit, "{$Endpoint}/user"},
		{Op(0), _Lbrace, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Authorization"},
		{Op(0), _Colon, LitType(0), ""},
		{Op(0), _Literal, StringLit, "Bearer {$Token}"},
		{Op(0), _Comma, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Content-Type"},
		{Op(0), _Colon, LitType(0), ""},
		{Op(0), _Literal, StringLit, "application/json; charset=utf-8"},
		{Op(0), _Comma, LitType(0), ""},
		{Op(0), _Rbrace, LitType(0), ""},
		{Op(0), _Arrow, LitType(0), ""},
		{Op(0), _Name, LitType(0), "send"},
		{Op(0), _Semi, LitType(0), ""},
		{Op(0), _Match, LitType(0), "match"},
		{Op(0), _Ref, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Resp"},
		{Op(0), _Dot, LitType(0), ""},
		{Op(0), _Name, LitType(0), "StatusCode"},
		{Op(0), _Lbrace, LitType(0), ""},
		{Op(0), _Literal, IntLit, "200"},
		{Op(0), _Arrow, LitType(0), ""},
		{Op(0), _Name, LitType(0), "print"},
		{Op(0), _Ref, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Resp"},
		{Op(0), _Dot, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Body"},
		{Op(0), _Ref, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Stdout"},
		{Op(0), _Semi, LitType(0), ""},
		{Op(0), _Name, LitType(0), "_"},
		{Op(0), _Arrow, LitType(0), ""},
		{Op(0), _Lbrace, LitType(0), ""},
		{Op(0), _Name, LitType(0), "print"},
		{Op(0), _Ref, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Resp"},
		{Op(0), _Dot, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Body"},
		{Op(0), _Ref, LitType(0), ""},
		{Op(0), _Name, LitType(0), "Stderr"},
		{Op(0), _Semi, LitType(0), ""},
		{Op(0), _Rbrace, LitType(0), ""},
		{Op(0), _Rbrace, LitType(0), ""},
		{Op(0), _EOF, LitType(0), ""},
	}

	for i, test := range tests {
		if sc.op != test.op {
			t.Fatalf("tests[%d] - unexpected op at %s, expected=%q, got=%q\n", i, sc.pos, test.op, sc.op)
		}
		if sc.tok != test.tok {
			t.Fatalf("tests[%d] - unexpected token at %s, expected=%q, got=%q\n", i, sc.pos, test.tok, sc.tok)
		}
		if sc.typ != test.typ {
			t.Fatalf("tests[%d] - unexpected type at %s, expected=%q, got=%q\n", i, sc.pos, test.typ, sc.typ)
		}
		if sc.lit != test.lit {
			t.Fatalf("tests[%d] - unexpected literal at %s, expected=%q, got=%q\n", i, sc.pos, test.lit, sc.lit)
		}

		t.Log(sc.pos, sc.tok, sc.typ, sc.lit)
		sc.next()
	}
}
