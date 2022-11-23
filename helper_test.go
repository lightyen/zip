package zip

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestCompress(t *testing.T) {
	output := "output.zip"
	password := []byte("golang")
	assets := []string{"./LICENSE", "README.md", "testdata"}

	f, err := os.Create(output)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	zw := NewWriter(f)
	defer zw.Close()

	err = WalkFiles(func(name, path string, d fs.DirEntry) error {
		if path == output {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		hdr, err := FileInfoHeader(info)
		if err != nil {
			return err
		}

		hdr.Name = name
		hdr.Method = Deflate

		ignored := "*/symlink.zip"
		matched, err := filepath.Match(ignored, name)
		if err != nil {
			return err
		}

		var w io.Writer
		switch {
		case matched:
			w, err = zw.CreateHeader(hdr)
		case len(password) > 0:
			w, err = zw.Encrypt(hdr, []byte(password))
		default:
			w, err = zw.CreateHeader(hdr)
		}

		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(w, f)
		return err
	}, assets...)

	if err != nil {
		t.Fatal(err)
	}
}

func TestExtract(t *testing.T) {
	password := []byte("golang")
	r, err := OpenReader("output.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	outputDir := "output"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	write := func(outputDir string, fi *File) (int64, error) {
		if fi.IsEncrypted() {
			fi.SetPassword(password)
		}

		r, err := fi.Open()
		if err != nil {
			return 0, err
		}
		defer r.Close()

		filename := filepath.Join(outputDir, fi.Name)
		if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			return 0, err
		}

		f, err := os.Create(filename)
		if err != nil {
			return 0, err
		}
		defer f.Close()

		return io.Copy(f, r)
	}

	for _, fi := range r.File {
		n, err := write(outputDir, fi)
		if err != nil {
			if err == ErrDecryption {
				t.Log(err)
				continue
			}
			t.Fatal(err)
		}

		t.Logf("Size of %v: %v byte(s)\n", filepath.Join(outputDir, fi.Name), n)
	}
}
