package eval

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/andrewpillar/req/syntax"
)

func errh(t *testing.T) func(pos syntax.Pos, msg string) {
	return func(pos syntax.Pos, msg string) {
		t.Errorf("%s - %s\n", pos, msg)
	}
}

func readfile(t *testing.T, fname string) io.Reader {
	b, err := os.ReadFile(fname)

	if err != nil {
		t.Fatal(err)
	}
	return strings.NewReader(strings.Replace(string(b), "__endpoint__", server.URL, -1))
}

var server *httptest.Server

func Test_Eval(t *testing.T) {
	ents, err := os.ReadDir("testdata")

	if err != nil {
		t.Fatal(err)
	}

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, ck := range r.Cookies() {
			println(ck.String())
			http.SetCookie(w, ck)
		}
		w.WriteHeader(http.StatusOK)
	}))

	var buf bytes.Buffer

	encodetab["form-data"].Func = encodeFormData("Test_Eval")

	for _, ent := range ents {
		if ent.IsDir() {
			continue
		}

		if !strings.HasSuffix(ent.Name(), ".req") {
			continue
		}

		fname := filepath.Join("testdata", ent.Name())
		out := fname[:len(fname)-3] + "out"

		nn, err := syntax.Parse(fname, readfile(t, fname), errh(t))

		if err != nil {
			t.Fatal(err)
		}

		e := New(&buf)

		if err := e.Run(nn); err != nil {
			t.Fatal(err)
		}

		b, err := os.ReadFile(out)

		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(buf.Bytes(), b) {
			t.Fatalf("unexpected output for script %q\n\n\texpected %q\n\n\t     got %q", fname, string(b), buf.String())
		}
		buf.Reset()
	}
}

func Test_Uuid(t *testing.T) {
	expr := `Uuid = uuid; write _ $Uuid;`
	nn, err := syntax.Parse("-", strings.NewReader(expr), errh(t))

	if err != nil {
		t.Fatalf("%s\n", err)
	}

	buf := &bytes.Buffer{}
	e := New(buf)

	err = e.Run(nn)

	if err != nil {
		t.Fatalf("expected evaluation of %q to be successful\n", expr)
	}

	uuidV4Format := `^[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}$`
	matched, err := regexp.MatchString(uuidV4Format, buf.String())

	if err != nil {
		t.Fatalf("could not parse regex %q\n", uuidV4Format)
	}
	if !matched {
		t.Fatalf("expected %q to match %q\n", buf.String(), uuidV4Format)
	}
}

func Test_EvalErrors(t *testing.T) {
	tests := []struct {
		expr string
		pos  syntax.Pos
	}{
		{`encode base64 "Hello world" -> command;`, syntax.Pos{Line: 1, Col: 32}},
		{`if "10" == 10 { }`, syntax.Pos{Line: 1, Col: 9}},
		{`Arr = []; writeln _ $Arr[true];`, syntax.Pos{Line: 1, Col: 25}},
		{`Arr = []; writeln _ $Arr["true"];`, syntax.Pos{Line: 1, Col: 25}},
		{`Arr = []; writeln _ "$(Arr["true"])";`, syntax.Pos{Line: 1, Col: 22}},
		{`writeln _ $Undefined;`, syntax.Pos{Line: 1, Col: 12}},
		{`writeln _ "Hello $(Undefined)";`, syntax.Pos{Line: 1, Col: 18}},
		{`if true { S = "block"; } writeln _ "S = $(S)";`, syntax.Pos{Line: 1, Col: 41}},
	}

	for i, test := range tests {
		nn, err := syntax.Parse("-", strings.NewReader(test.expr), errh(t))

		if err != nil {
			t.Fatalf("tests[%d] - %s\n", i, err)
		}

		e := New(os.Stdout)

		err = e.Run(nn)

		if err == nil {
			t.Fatalf("tests[%d] - expected evaluation of %q to error\n", i, test.expr)
		}

		evalerr, ok := err.(Error)

		if !ok {
			t.Fatalf("tests[%d] - unexpected error type, expected=%T, got=%T(%q)\n", i, Error{}, err, err)
		}

		if test.pos.Line != evalerr.Pos.Line || test.pos.Col != evalerr.Pos.Col {
			t.Fatalf("tests[%d] - unexpected error position, expected=%q, got=%q\n", i, test.pos, evalerr.Pos)
		}
	}
}
