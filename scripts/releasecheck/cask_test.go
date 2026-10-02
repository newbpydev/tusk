package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func caskFixture() manifest {
	m := manifest{Version: "0.3.0", Mode: "candidate", SourceSHA: strings.Repeat("a", 40), Files: map[string]string{}}
	for _, arch := range []string{"amd64", "arm64"} {
		target := "darwin/" + arch
		members := map[string]string{"tusk": strings.Repeat("a", 64), "completions/tusk.bash": "hash", "completions/tusk.fish": "hash", "completions/tusk.zsh": "hash", "man/tusk.1": "hash"}
		name := archiveName(m.Version, target)
		m.Files[name] = strings.Repeat("a", 64)
		m.Targets = append(m.Targets, targetRecord{Target: target, Archive: name, Members: members, ExecutableSHA256: members["tusk"]})
	}
	return m
}
func TestCask_IdentityAndDeclarativeBoundary(t *testing.T) {
	m := caskFixture()
	text, err := caskText(m)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Index(text, []byte("on_arm do")) > bytes.Index(text, []byte("on_intel do")) {
		t.Fatal("pinned packager places ARM before Intel")
	}
	for _, expected := range []string{"on_intel do", "on_arm do", "depends_on :macos", "binary \"tusk\"", "manpage \"man/tusk.1\"", "bash_completion", "zsh_completion", "fish_completion"} {
		if !bytes.Contains(text, []byte(expected)) {
			t.Fatal(expected)
		}
	}
	for _, mutate := range []func(*manifest){
		func(m *manifest) { m.Version = "0.3.0\";system(\"evil\")" },
		func(m *manifest) { m.Mode = "snapshot" },
		func(m *manifest) { m.Targets = m.Targets[:1] },
		func(m *manifest) { m.Targets = append(m.Targets, m.Targets[0]) },
		func(m *manifest) { m.Targets[0].Archive = "../../db" },
		func(m *manifest) { m.Files[m.Targets[0].Archive] = "bad" },
		func(m *manifest) { delete(m.Targets[0].Members, "completions/tusk.bash") },
		func(m *manifest) { m.Targets[1].Members["man/evil\".1"] = "hash" },
		func(m *manifest) { delete(m.Targets[1].Members, "man/tusk.1") },
		func(m *manifest) { delete(m.Targets[0].Members, "tusk") },
		func(m *manifest) { m.Targets[0].ExecutableSHA256 = strings.Repeat("b", 64) },
	} {
		m := caskFixture()
		mutate(&m)
		if _, err := caskText(m); err == nil {
			t.Fatal("unsafe manifest accepted")
		}
	}
	dir := t.TempDir()
	mf := filepath.Join(dir, "release-manifest.json")
	file := filepath.Join(dir, "tusk.rb")
	if err := writeJSON(mf, caskFixture()); err != nil {
		t.Fatal(err)
	}
	if run([]string{"render-cask", mf, file}, os.Stderr) != 0 {
		t.Fatal("render")
	}
	if run([]string{"render-cask", mf, file}, os.Stderr) == 0 {
		t.Fatal("output overwritten")
	}
	if run([]string{"check-cask", mf, file}, os.Stderr) != 0 {
		t.Fatal("check")
	}
	for _, change := range []string{string(text) + "zap trash: '~/.local/share/tusk'\n", strings.Replace(string(text), "on_intel do", "on_arm do", 1), strings.Replace(string(text), strings.Repeat("a", 64), strings.Repeat("b", 64), 1), strings.Replace(string(text), "\n    end\n", "\n    oops\n", 1), string(text) + "postflight do; system_command '/usr/bin/xattr'; end\n"} {
		if err := os.WriteFile(file, []byte(change), 0644); err != nil {
			t.Fatal(err)
		}
		if run([]string{"check-cask", mf, file}, os.Stderr) == 0 {
			t.Fatal("changed cask accepted")
		}
	}
	if run([]string{"check-cask", mf, filepath.Join(dir, "missing")}, os.Stderr) == 0 {
		t.Fatal("missing cask accepted")
	}
	if run([]string{"render-cask", filepath.Join(dir, "missing"), filepath.Join(dir, "new")}, os.Stderr) == 0 {
		t.Fatal("missing manifest accepted")
	}
	if run([]string{"render-cask", mf, filepath.Join(dir, "absent", "new")}, os.Stderr) == 0 {
		t.Fatal("missing parent accepted")
	}
}
