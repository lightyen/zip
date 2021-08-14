package zip

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompress(t *testing.T) {
	output := "output.zip"
	f, _ := os.Create(output)

	ignore, err := filepath.Glob(output)
	if err != nil {
		t.Fatal(err)
	}

	if err := Compress("golang", f, ignore, "./LICENSE", "README.md", "testdata"); err != nil {
		t.Fatal(err)
	}
}
