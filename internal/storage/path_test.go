package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePath_Precedence(t *testing.T) {
	base := t.TempDir()
	for _, tc := range []struct{ explicit, env, xdg, home, want string }{
		{"option.db", "env.db", "", "", filepath.Join(base, "option.db")},
		{"", "relative.db", "", "", filepath.Join(base, "relative.db")},
		{"", "", filepath.Join(base, "xdg"), "", filepath.Join(base, "xdg", "tusk", "tusk.db")},
		{"", "", "relative-xdg", base, filepath.Join(base, ".local", "share", "tusk", "tusk.db")},
		{"", "", "", base, filepath.Join(base, ".local", "share", "tusk", "tusk.db")},
	} {
		inputs := pathInputs{lookup: func(key string) string {
			if key == "TUSK_DB_PATH" {
				return tc.env
			}
			return tc.xdg
		}, home: func() (string, error) { return tc.home, nil }, cwd: func() (string, error) { return base, nil }}
		got, err := resolvePath(tc.explicit, inputs)
		if err != nil || got != tc.want {
			t.Fatalf("path=%s want=%s err=%v", got, tc.want, err)
		}
	}
	bad := func() (string, error) { return "", errors.New("lookup failed") }
	inputs := pathInputs{lookup: func(string) string { return "" }, home: bad, cwd: bad}
	abs := filepath.Join(base, "absolute.db")
	if got, err := resolvePath(abs, inputs); err != nil || got != abs {
		t.Fatalf("irrelevant fallback failed: %s %v", got, err)
	}
	for _, p := range []string{"", "relative.db", "bad\x00path", "bad\xffpath"} {
		if _, err := resolvePath(p, inputs); err == nil {
			t.Fatalf("accepted %q", p)
		}
	}
	inputs.home = func() (string, error) { return "relative-home", nil }
	if _, err := resolvePath("", inputs); err == nil {
		t.Fatal("relative home accepted")
	}
	inputs.cwd = func() (string, error) { return "relative-cwd", nil }
	if _, err := resolvePath("relative.db", inputs); err == nil {
		t.Fatal("relative cwd accepted")
	}
}

func TestPrepareFile_RejectsCreatedDevice(t *testing.T) {
	var opened *os.File
	err := prepareFileWith(filepath.Join(t.TempDir(), "device.db"), func(string, int, os.FileMode) (*os.File, error) {
		var err error
		opened, err = os.OpenFile(os.DevNull, os.O_RDWR, 0)
		return opened, err
	})
	if opened == nil {
		t.Fatal("creation fault was not exercised")
	}
	defer opened.Close()
	if err == nil {
		t.Fatal("created device accepted as regular database file")
	}
	if _, err := opened.Stat(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("rejected handle leaked: %v", err)
	}
}
