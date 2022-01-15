package eval

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

		PrintCmd.Func = print(&buf)

		nn, err := syntax.Parse(fname, readfile(t, fname), errh(t))

		if err != nil {
			t.Fatal(err)
		}

		e := New()

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
