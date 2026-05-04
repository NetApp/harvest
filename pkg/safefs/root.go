package safefs

import (
	"io/fs"
	"os"
	"path/filepath"
)

func WithRoot(dir string, fn func(root *os.Root) error) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	return fn(root)
}

func ReadFile(name string) ([]byte, error) {
	var data []byte
	err := WithRoot(filepath.Dir(name), func(root *os.Root) error {
		var err error
		data, err = root.ReadFile(filepath.Base(name))
		return err
	})
	return data, err
}

func WriteFile(name string, data []byte, perm os.FileMode) error {
	return WithRoot(filepath.Dir(name), func(root *os.Root) error {
		return root.WriteFile(filepath.Base(name), data, perm)
	})
}

func WalkDir(dir string, fn func(root *os.Root, path string, d fs.DirEntry) error) error {
	return WithRoot(dir, func(root *os.Root) error {
		return fs.WalkDir(root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			return fn(root, path, d)
		})
	})
}
