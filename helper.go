package zip

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func Compress(password string, dst io.Writer, ignore []string, assets ...string) error {
	for i := range ignore {
		ignore[i] = filepath.Clean(ignore[i])
		var err error
		ignore[i], err = filepath.Abs(ignore[i])
		if err != nil {
			return err
		}
	}

	zw := NewWriter(dst)
	defer zw.Close()

	type Item struct {
		IsDir bool
		Name  string
		Info  fs.FileInfo
	}

	var valid []Item

	clean := make(map[string]struct{})
	for i := range assets {
		clean[filepath.Clean(assets[i])] = struct{}{}
	}

	var dirs []string
	for name := range clean {
		info, err := os.Stat(name)
		if err != nil {
			return err
		}

		switch {
		case info.IsDir():
			valid = append(valid, Item{IsDir: true, Name: name, Info: info})
			r, err := filepath.Abs(name)
			if err != nil {
				return err
			}
			dirs = append(dirs, r)
		case info.Mode().IsRegular():
			valid = append(valid, Item{IsDir: false, Name: name, Info: info})
			r, err := filepath.Abs(name)
			if err != nil {
				return err
			}
			dirs = append(dirs, filepath.Dir(r))
		}
	}

	if len(valid) == 0 {
		return nil
	}

	longestCommonDir := func(dirs []string) string {
		if len(dirs) == 0 {
			return ""
		}

		if len(dirs) == 1 {
			return dirs[0]
		}

		c := dirs[0] + string(filepath.Separator)

		for _, v := range dirs[1:] {
			v = v + string(filepath.Separator)

			if len(v) < len(c) {
				c = c[:len(v)]
			}

			for i := 0; i < len(c); i++ {
				if v[i] != c[i] {
					c = c[:i]
					break
				}
			}
		}

		for i := len(c) - 1; i >= 0; i-- {
			if c[i] == filepath.Separator {
				c = c[:i]
				break
			}
		}

		return c
	}

	root := longestCommonDir(dirs)

	compress := func(name, path, password string, info fs.FileInfo) error {
		switch {
		case info.Mode().IsDir():
			if !strings.HasSuffix(name, "/") {
				name = name + "/"
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

		_, err = io.Copy(w, file)
		return err
	}

	for i := range valid {
		if valid[i].IsDir {
			err := filepath.WalkDir(valid[i].Name, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.Name() == "." {
					return nil
				}

				p, err := filepath.Abs(path)
				if err != nil {
					return err
				}

				for i := range ignore {
					if p == ignore[i] {
						return nil
					}
				}

				rel, err := filepath.Rel(root, p)
				if err != nil {
					return err
				}
				if rel == "." {
					return nil
				}

				name := filepath.ToSlash(rel)

				info, err := d.Info()
				if err != nil {
					return err
				}

				return compress(name, path, password, info)
			})

			if err != nil {
				return err
			}
		} else {
			do := func(path string) error {
				a, err := filepath.Abs(path)
				if err != nil {
					return err
				}

				rel, err := filepath.Rel(root, a)
				if err != nil {
					return err
				}

				name := filepath.ToSlash(rel)

				info, err := os.Stat(path)
				if err != nil {
					return err
				}

				return compress(name, path, password, info)
			}

			err := do(valid[i].Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
