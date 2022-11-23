package zip

import (
	"io/fs"
	"os"
	"path/filepath"
)

type WalkFilesFunc = func(name, path string, d fs.DirEntry) error

func WalkFiles(handler WalkFilesFunc, assets ...string) error {
	type Item struct {
		IsDir bool
		Name  string
		Info  fs.FileInfo
	}

	var items []Item

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
			items = append(items, Item{IsDir: true, Name: name, Info: info})
			r, err := filepath.Abs(name)
			if err != nil {
				return err
			}
			dirs = append(dirs, r)
		case info.Mode().IsRegular():
			items = append(items, Item{IsDir: false, Name: name, Info: info})
			r, err := filepath.Abs(name)
			if err != nil {
				return err
			}
			dirs = append(dirs, filepath.Dir(r))
		}
	}

	if len(items) == 0 {
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

	for i := range items {
		if items[i].IsDir {
			err := filepath.WalkDir(items[i].Name, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if !d.Type().IsRegular() {
					return nil
				}

				abs, err := filepath.Abs(path)
				if err != nil {
					return err
				}

				rel, err := filepath.Rel(root, abs)
				if err != nil {
					return err
				}

				name := filepath.ToSlash(rel)

				return handler(name, path, d)
			})

			if err != nil {
				return err
			}
		} else {
			do := func(path string) error {
				abs, err := filepath.Abs(path)
				if err != nil {
					return err
				}

				rel, err := filepath.Rel(root, abs)
				if err != nil {
					return err
				}

				name := filepath.ToSlash(rel)

				info, err := os.Stat(path)
				if err != nil {
					return err
				}

				return handler(name, path, &statDirEntry{info})
			}

			if err := do(items[i].Name); err != nil && err != fs.SkipDir {
				return err
			}
		}
	}

	return nil
}

type statDirEntry struct {
	info fs.FileInfo
}

func (d *statDirEntry) Name() string               { return d.info.Name() }
func (d *statDirEntry) IsDir() bool                { return d.info.IsDir() }
func (d *statDirEntry) Type() fs.FileMode          { return d.info.Mode().Type() }
func (d *statDirEntry) Info() (fs.FileInfo, error) { return d.info, nil }
