package zip

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Compress(password string, dst io.Writer, filename string, assets ...string) error {
	zw := NewWriter(dst)
	defer zw.Close()

	if len(assets) == 0 {
		return nil
	}

	skip := true
	info, err := os.Stat(assets[0])
	if err != nil {
		return err
	}
	if !info.IsDir() {
		skip = false
	}

	if len(assets) > 1 {
		skip = false
	}

	root := ""
	a, _ := filepath.Abs(filename)
	b := filepath.Base(filename)

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if b == filepath.Base(path) {
			p, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			if a == p {
				return nil
			}
		}

		var name string
		if skip {
			rel, err := filepath.Rel(assets[0], path)
			if err != nil {
				return err
			}
			if rel == "." {
				return nil // skip
			}
			name = filepath.ToSlash(rel)
		} else {
			name = filepath.ToSlash(filepath.Join(filepath.Base(root), strings.TrimPrefix(path, root)))
		}

		switch {
		case info.Mode().IsDir():
			if !strings.HasSuffix(name, "/") {
				name = name + "/"
			}
		case info.Mode()&os.ModeSymlink == os.ModeSymlink:
			info, err = os.Stat(path)
			if err != nil {
				return err
			}
		case !info.Mode().IsRegular():
			return nil
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

		if !info.Mode().IsRegular() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err = io.Copy(w, file); err != nil {
			return err
		}
		return nil
	}

	for _, a := range assets {
		a = filepath.Clean(a)
		info, err := os.Stat(a)
		if err != nil {
			return err
		}
		root = a
		if info.IsDir() {
			if err := filepath.Walk(a, walk); err != nil {
				return err
			}
		} else if err := walk(a, info, err); err != nil {
			return err
		}
	}
	return nil
}
