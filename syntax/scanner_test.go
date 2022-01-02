package syntax

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrewpillar/req/token"
)

func errh(t *testing.T) func(pos token.Pos, msg string) {
	return func(pos token.Pos, msg string) {
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
		tok token.Token
		typ token.Type
		lit string
	}{
		{token.Name, token.Type(0), "Stdout"},
		{token.Assign, token.Type(0), ""},
		{token.Name, token.Type(0), "open"},
		{token.Literal, token.String, "/dev/stdout"},
		{token.Semi, token.Type(0), ""},
		{token.Name, token.Type(0), "Stderr"},
		{token.Assign, token.Type(0), ""},
		{token.Name, token.Type(0), "open"},
		{token.Literal, token.String, "/dev/stderr"},
		{token.Semi, token.Type(0), ""},
		{token.Name, token.Type(0), "Endpoint"},
		{token.Assign, token.Type(0), ""},
		{token.Literal, token.String, "https://api.github.com"},
		{token.Semi, token.Type(0), ""},
		{token.Name, token.Type(0), "Token"},
		{token.Assign, token.Type(0), ""},
		{token.Name, token.Type(0), "env"},
		{token.Literal, token.String, "GH_TOKEN"},
		{token.Semi, token.Type(0), ""},
		{token.Name, token.Type(0), "Resp"},
		{token.Assign, token.Type(0), ""},
		{token.Name, token.Type(0), "GET"},
		{token.Literal, token.String, "{$Endpoint}/user"},
		{token.Lbrace, token.Type(0), ""},
		{token.Name, token.Type(0), "Authorization"},
		{token.Colon, token.Type(0), ""},
		{token.Literal, token.String, "Bearer {$Token}"},
		{token.Comma, token.Type(0), ""},
		{token.Name, token.Type(0), "Content-Type"},
		{token.Colon, token.Type(0), ""},
		{token.Literal, token.String, "application/json; charset=utf-8"},
		{token.Comma, token.Type(0), ""},
		{token.Rbrace, token.Type(0), ""},
		{token.Arrow, token.Type(0), ""},
		{token.Name, token.Type(0), "send"},
		{token.Semi, token.Type(0), ""},
		{token.Name, token.Type(0), "print"},
		{token.Ref, token.Type(0), ""},
		{token.Name, token.Type(0), "Resp"},
		{token.Dot, token.Type(0), ""},
		{token.Name, token.Type(0), "Body"},
		{token.Match, token.Type(0), "match"},
		{token.Ref, token.Type(0), ""},
		{token.Name, token.Type(0), "Resp"},
		{token.Dot, token.Type(0), ""},
		{token.Name, token.Type(0), "StatusCode"},
		{token.Lbrace, token.Type(0), ""},
		{token.Literal, token.Int, "200"},
		{token.Arrow, token.Type(0), ""},
		{token.Yield, token.Type(0), "yield"},
		{token.Ref, token.Type(0), ""},
		{token.Name, token.Type(0), "Stdout"},
		{token.Comma, token.Type(0), ""},
		{token.Name, token.Type(0), "_"},
		{token.Arrow, token.Type(0), ""},
		{token.Lbrace, token.Type(0), ""},
		{token.Yield, token.Type(0), "yield"},
		{token.Ref, token.Type(0), ""},
		{token.Name, token.Type(0), "Stderr"},
		{token.Semi, token.Type(0), ""},
		{token.Rbrace, token.Type(0), ""},
		{token.Comma, token.Type(0), ""},
		{token.Rbrace, token.Type(0), ""},
		{token.EOF, token.Type(0), ""},
	}

	for i, test := range tests {
		if sc.tok != test.tok {
			t.Fatalf("tests[%d] - unexpected token, expected=%q, got=%q\n", i, test.tok, sc.tok)
		}
		if sc.typ != test.typ {
			t.Fatalf("tests[%d] - unexpected type, expected=%q, got=%q\n", i, test.typ, sc.typ)
		}
		if sc.lit != test.lit {
			t.Fatalf("tests[%d] - unexpected literal, expected=%q, got=%q\n", i, test.lit, sc.lit)
		}

		t.Log(sc.pos, sc.tok, sc.typ, sc.lit)
		sc.next()
	}
}
