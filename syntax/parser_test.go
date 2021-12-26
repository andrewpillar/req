package syntax

import (
	"path/filepath"
	"testing"
)

func Test_Parser(t *testing.T) {
	_, err := ParseFile(filepath.Join("testdata", "gh.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}
}
