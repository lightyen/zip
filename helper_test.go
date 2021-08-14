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
	ignored, err := filepath.Glob(output)
	if err != nil {
		t.Fatal(err)
	}

	password := "golang"
	assets := []string{"./LICENSE", "README.md", "testdata"}

	f, _ := os.Create(output)
	zw := NewWriter(f)
	defer zw.Close()

	err = WalkFiles(func(name, path string, d fs.DirEntry) error {
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

		var w io.Writer

		if password != "" {
			w, err = zw.Encrypt(hdr, password)
		} else {
			w, err = zw.CreateHeader(hdr)
		}

		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		n, err := io.Copy(w, file)
		if err == nil {
			t.Logf("Size of %v: %v byte(s)\n", path, n)
		}
		return err
	}, ignored, assets...)

	if err != nil {
		t.Fatal(err)
	}
}

func TestExtract(t *testing.T) {
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
			fi.SetPassword("golang")
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
