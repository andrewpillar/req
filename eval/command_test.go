package eval

//import (
//	"bytes"
//	"io"
//	"net/http"
//	"os"
//	"path/filepath"
//	"testing"
//
//	"github.com/andrewpillar/req/value"
//)
//
//func Test_CommandError(t *testing.T) {
//	cmd := &Command{
//		Name: "test-cmd",
//		Argc: 1,
//	}
//
//	tests := []struct {
//		args []value.Value
//		err  error
//	}{
//		{[]value.Value{}, errNotEnoughArgs},
//		{[]value.Value{value.String{}, value.Int{}}, errTooManyArgs},
//	}
//
//	for i, test := range tests {
//		_, err := cmd.invoke(test.args)
//
//		cmderr, ok := err.(*CommandError)
//
//		if !ok {
//			t.Errorf("tests[%d] - unexpected error type, expected=%T, got=%T\n", i, &CommandError{}, err)
//			continue
//		}
//
//		if cmderr.Err != test.err {
//			t.Errorf("tests[%d] - unexpected error %q, got %T(%q)\n", i, cmderr.Err, test.err, test.err)
//		}
//	}
//}
//
//func Test_EnvCmd(t *testing.T) {
//	key := value.String{
//		Value: "TOKEN",
//	}
//
//	val, err := EnvCmd.invoke([]value.Value{key})
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	_, err = value.ToString(val)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	tok := "123456"
//	os.Setenv(key.Value, tok)
//
//	val, err = EnvCmd.invoke([]value.Value{key})
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	str, err := value.ToString(val)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if str.Value != tok {
//		t.Fatalf("unexpected TOKEN value, expected=%q, got=%q\n", tok, str.Value)
//	}
//}
//
//func Test_OpenCmd(t *testing.T) {
//	fname := value.String{
//		Value: filepath.Join("testdata", "example"),
//	}
//
//	val, err := OpenCmd.invoke([]value.Value{fname})
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	f, err := value.ToFile(val)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	f.Close()
//}
//
//func Test_PrintCmd(t *testing.T) {
//	f, err := os.CreateTemp("", "req-Test_PrintCmd")
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	defer os.RemoveAll(f.Name())
//
//	args := []value.Value{
//		&value.Array{},
//		&value.FormData{},
//		value.Int{},
//		value.Name{},
//		value.Object{},
//		value.Request{},
//		value.Response{},
//		value.String{},
//		value.Zero{},
//		value.File{File: f},
//	}
//
//	_, err = PrintCmd.invoke(args)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if _, err := f.Seek(0, io.SeekStart); err != nil {
//		t.Fatal(err)
//	}
//
//	b, err := io.ReadAll(f)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	expected := "[]  0  {}     \n"
//
//	if string(b) != expected {
//		t.Errorf("print result did not match, expected=%q, got=%q\n", expected, string(b))
//	}
//}
//
//func Test_RequestCmd(t *testing.T) {
//	tests := []struct {
//		cmd    string
//		args   []value.Value
//		url    string
//		header http.Header
//		body   io.ReadCloser
//	}{
//		{
//			"POST",
//			[]value.Value{
//				value.String{Value: "https://example.com"},
//			},
//			"https://example.com",
//			nil,
//			nil,
//		},
//		{
//			"POST",
//			[]value.Value{
//				value.String{Value: "https://example.com"},
//				value.Object{
//					Pairs: map[string]value.Value{
//						"Content-Type": value.String{
//							Value: "application/json",
//						},
//					},
//				},
//			},
//			"https://example.com",
//			http.Header{
//				"Content-Type": {"application/json"},
//			},
//			nil,
//		},
//		{
//			"POST",
//			[]value.Value{
//				value.String{Value: "https://example.com"},
//				value.Object{
//					Pairs: map[string]value.Value{
//						"Content-Type": value.String{
//							Value: "application/json",
//						},
//					},
//				},
//				value.String{
//					Value: `{"username": "admin", "password": "secret"}`,
//				},
//			},
//			"https://example.com",
//			http.Header{
//				"Content-Type": {"application/json"},
//			},
//			io.NopCloser(bytes.NewBufferString(`{"username": "admin", "password": "secret"}`)),
//		},
//	}
//
//	for i, test := range tests {
//		val, err := request(test.cmd, test.args)
//
//		if err != nil {
//			t.Errorf("tests[%d] - %s\n", i, err)
//			continue
//		}
//
//		req, err := value.ToRequest(val)
//
//		if err != nil {
//			t.Errorf("tests[%d] - %s\n", i, err)
//			continue
//		}
//
//		if req.Method != test.cmd {
//			t.Errorf("tests[%d] - unexpected request method, expected=%q, got=%q\n", i, test.cmd, req.Method)
//			continue
//		}
//
//		if test.header != nil {
//			for k, v := range test.header {
//				v2, ok := req.Header[k]
//
//				if !ok {
//					t.Errorf("tests[%d] - expected header %q in request\n", i, k)
//					continue
//				}
//
//				if v[0] != v2[0] {
//					t.Errorf("tests[%d] - request header does not match, expected=%q, got=%q\n", i, v[0], v2[0])
//				}
//			}
//		}
//
//		if test.body != nil {
//			b, _ := io.ReadAll(test.body)
//			b2, _ := io.ReadAll(req.Body)
//
//			if !bytes.Equal(b, b2) {
//				t.Errorf("tests[%d] - request body does not match, expected=%q, got=%q\n", i, string(b), string(b2))
//			}
//		}
//	}
//}
//
//func Test_SniffCmd(t *testing.T) {
//	f, err := os.Open(filepath.Join("testdata", "example.html"))
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	defer f.Close()
//
//	val, err := SniffCmd.invoke([]value.Value{value.File{File: f}})
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	str, err := value.ToString(val)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	mime := "text/html; charset=utf-8"
//
//	if str.Value != mime {
//		t.Errorf("unexpected mime, expected=%q, got=%q\n", mime, str.Value)
//	}
//}
