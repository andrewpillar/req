package eval

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrewpillar/req/syntax"
	"github.com/andrewpillar/req/value"
)

var server *httptest.Server

func errh(t *testing.T) func(syntax.Pos, string) {
	return func(pos syntax.Pos, msg string) {
		t.Errorf("%s - %s\n", pos, msg)
	}
}

func readfile(t *testing.T, fname string) io.Reader {
	b, err := os.ReadFile(filepath.Join("testdata", fname))

	if err != nil {
		t.Fatal(err)
	}
	return strings.NewReader(strings.Replace(string(b), "__endpoint__", server.URL, -1))
}

func Test_EvalVarDecl(t *testing.T) {
	nn, err := syntax.Parse("vardecl.req", readfile(t, "vardecl.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	e := New()

	var c context

	for _, n := range nn {
		if _, err := e.eval(&c, n); err != nil {
			t.Errorf("%s\n", err)
		}
	}

	tests := []struct {
		varname string
		typname string
	}{
		{"String", "string"},
		{"Number", "int"},
		{"Bool", "bool"},
		{"Array", "array"},
		{"Object", "object"},
	}

	for i, test := range tests {
		val, err := c.Get(test.varname)

		if err != nil {
			t.Errorf("tests[%d] - %s\n", i, err)
			continue
		}

		if typ := value.Type(val); typ != test.typname {
			t.Errorf("tests[%d] - unexpected type for variable %q, expected=%q, got=%q\n", i, test.varname, test.typname, typ)
		}
	}
}

func Test_EvalRef(t *testing.T) {
	nn, err := syntax.Parse("refexpr.req", readfile(t, "refexpr.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	e := New()

	if err := e.Run(nn); err != nil {
		t.Fatal(err)
	}
}

func Test_EvalInterpolate(t *testing.T) {
	nn, err := syntax.Parse("vardecl.req", readfile(t, "vardecl.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	e := New()

	var c context

	for _, n := range nn {
		if _, err := e.eval(&c, n); err != nil {
			t.Errorf("%s\n", err)
		}
	}

	tests := []struct {
		input    string
		expected string
	}{
		{`{$String}`, "string"},
		{`{$Number}`, "10"},
		{`{$Bool}`, "true"},
		{`{$Array[0]}`, "1"},
		{`{$Array[2]}`, "3"},
		{`{$Array[3]}`, "4"},
		{`{$Array}`, "[1 2 3 4]"},
		{`{$Object["String"]}`, "string"},
		{`{$Object["Array"][0]}`, "1"},
		{`{$Object["Child"]["Array"][2]}`, "three"},
		{`{$Object}`, "{Array:[1 2 3 4] Child:{Array:[one two three] Number:10} String:string}"},
	}

	for i, test := range tests {
		val, err := e.interpolate(&c, test.input)

		if err != nil {
			t.Errorf("tests[%d] - failed to interpolate string: %s\n", i, err)
			continue
		}

  		str, err := value.ToString(val)

		if err != nil {
			t.Fatalf("tests[%d] - Eval.interpolate did not return a stringObj", i)
		}

		if str.Value != test.expected {
			t.Errorf("tests[%d] - unexpected output for %q, expected=%q, got=%q\n", i, test.input, test.expected, str.Value)
		}
	}
}

func TestMain(m *testing.M) {
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	code := m.Run()

	if code != 0 {
		os.Exit(code)
	}
}
