package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type pathInputs struct {
	lookup func(string) string
	home   func() (string, error)
	cwd    func() (string, error)
}

// Resolve returns the absolute database path for an explicit path selection,
// applying the documented TUSK_DB_PATH -> absolute XDG_DATA_HOME -> home
// fallback precedence when explicit is empty. It reads the environment and
// working directory but performs no filesystem access and opens nothing;
// composition roots use it to pin one path up front instead of duplicating
// the precedence rules.
func Resolve(explicit string) (string, error) {
	return resolvePath(explicit, pathInputs{lookup: os.Getenv, home: os.UserHomeDir, cwd: os.Getwd})
}

func resolvePath(explicit string, in pathInputs) (string, error) {
	path := explicit
	if path == "" {
		path = in.lookup("TUSK_DB_PATH")
	}
	if path == "" {
		xdg := in.lookup("XDG_DATA_HOME")
		if filepath.IsAbs(xdg) {
			path = filepath.Join(xdg, "tusk", "tusk.db")
		} else {
			home, err := in.home()
			if err != nil {
				return "", err
			}
			if !filepath.IsAbs(home) {
				return "", fmt.Errorf("storage: home must be absolute")
			}
			path = filepath.Join(home, ".local", "share", "tusk", "tusk.db")
		}
	}
	if !utf8.ValidString(path) || strings.ContainsRune(path, 0) {
		return "", fmt.Errorf("storage: invalid database path")
	}
	if !filepath.IsAbs(path) {
		cwd, err := in.cwd()
		if err != nil {
			return "", err
		}
		if !filepath.IsAbs(cwd) {
			return "", fmt.Errorf("storage: working directory must be absolute")
		}
		path = filepath.Join(cwd, path)
	}
	return filepath.Clean(path), nil
}

func prepareFile(path string) error {
	return prepareFileWith(path, os.OpenFile)
}

func prepareFileWith(path string, create func(string, int, os.FileMode) (*os.File, error)) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}
		f, createErr := create(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if createErr == nil {
			info, err = f.Stat()
			err = errors.Join(err, f.Close())
		} else {
			if !errors.Is(createErr, os.ErrExist) {
				return createErr
			}
			info, err = os.Lstat(path)
		}
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || isReparse(info) {
		return fmt.Errorf("storage: database target must be a regular file")
	}
	return nil
}
