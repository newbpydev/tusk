package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

func TestOverlay_ImmutableAndIsolated(t *testing.T) {
	root := filepath.Join(t.TempDir(), "source ü")
	if err := os.MkdirAll(filepath.Join(root, "cmd", "tusk"), 0700); err != nil {
		t.Fatal(err)
	}
	original := []byte("package main\nconst Version = \"dev\"\n")
	versionFile := filepath.Join(root, "cmd", "tusk", "version.go")
	if err := os.WriteFile(versionFile, original, 0600); err != nil {
		t.Fatal(err)
	}
	versionFile, err := filepath.EvalSymlinks(versionFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, snapshot := range []bool{false, true} {
		destination := filepath.Join(t.TempDir(), "overlay ü")
		info, err := createOverlay(root, destination, "v0.3.0", snapshot)
		if err != nil {
			t.Fatal(err)
		}
		var overlay struct{ Replace map[string]string }
		data, err := os.ReadFile(filepath.Join(destination, "overlay.json"))
		if err != nil || json.Unmarshal(data, &overlay) != nil {
			t.Fatalf("overlay: %v", err)
		}
		generated, err := os.ReadFile(overlay.Replace[versionFile])
		if err != nil || !strings.Contains(string(generated), `const Version = "`+info.Version+`"`) || digest(generated) != info.InputSHA256 {
			t.Fatalf("constant input: %s %v", generated, err)
		}
		want := "0.3.0"
		if snapshot {
			want += "-dev"
		}
		if info.Version != want {
			t.Fatal(info)
		}
		if _, err := createOverlay(root, destination, "v0.3.0", snapshot); err == nil {
			t.Fatal("accepted output collision")
		}
	}
	data, err := os.ReadFile(versionFile)
	if err != nil || !bytes.Equal(data, original) {
		t.Fatal("original source changed")
	}
	for _, version := range []string{"", "v1", "v01.2.3", "v1.2.3\";panic()", "v1.2.3$(touch x)", "v1.2.3-dev"} {
		if _, err := createOverlay(root, filepath.Join(t.TempDir(), "out"), version, false); err == nil {
			t.Fatalf("invalid version accepted: %q", version)
		}
	}
	if _, err := createOverlay(root, filepath.Join(root, "overlay"), "v0.3.0", false); err == nil {
		t.Fatal("overlay entered source checkout")
	}
}

func TestReleaseCLI_OverlayAndErrors(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "tusk"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "tusk", "version.go"), []byte("package main\nconst Version=\"dev\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "overlay")
	for _, args := range [][]string{nil, {"unknown"}, {"overlay"}, {"overlay", root, output, "vbad", "candidate"}, {"overlay", root, output, "v0.3.0", "invalid"}} {
		if run(args, io.Discard) == 0 {
			t.Fatalf("accepted arguments %q", args)
		}
	}
	if run([]string{"overlay", root, output, "v0.3.0", "candidate"}, io.Discard) != 0 {
		t.Fatal("overlay CLI failed")
	}
	if data, err := os.ReadFile(filepath.Join(output, "input.json")); err != nil || !bytes.Contains(data, []byte(`"version": "0.3.0"`)) {
		t.Fatalf("missing input receipt: %s %v", data, err)
	}
}

func tarFixture(t *testing.T, headers []*tar.Header, contents []string) string {
	path := filepath.Join(t.TempDir(), "fixture.tar.gz")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	tarOut := tar.NewWriter(gz)
	for index, header := range headers {
		if err := tarOut.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if header.Typeflag == tar.TypeReg && header.Size <= int64(len(contents[index])) {
			if _, err := tarOut.Write([]byte(contents[index])); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tarOut.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestArchive_RejectUnsafeMembers(t *testing.T) {
	for _, mode := range []int64{04755, 02755, 01755} {
		path := tarFixture(t, []*tar.Header{{Name: "tusk", Mode: mode, Size: 1, Typeflag: tar.TypeReg}}, []string{"x"})
		if _, err := readArchive(path); err == nil {
			t.Fatalf("special permission bits accepted: %o", mode)
		}
	}
	for _, name := range []string{"/absolute", "../escape", "a/../escape", "C:/escape", `a\escape`, "a/.git/config", ".env", "tasks.db", "tasks.db-wal", "bad\x1bname"} {
		path := tarFixture(t, []*tar.Header{{Name: name, Mode: 0644, Size: 1, Typeflag: tar.TypeReg}}, []string{"x"})
		if _, err := readArchive(path); err == nil {
			t.Fatalf("unsafe member accepted: %q", name)
		}
	}
	for _, kind := range []byte{tar.TypeSymlink, tar.TypeLink, tar.TypeFifo} {
		path := tarFixture(t, []*tar.Header{{Name: "link", Linkname: "outside", Mode: 0644, Typeflag: kind}}, []string{""})
		if _, err := readArchive(path); err == nil {
			t.Fatal("unsafe member type accepted")
		}
	}
	path := tarFixture(t, []*tar.Header{{Name: "README.md", Mode: 0644, Size: 1, Typeflag: tar.TypeReg}, {Name: "README.md", Mode: 0644, Size: 1, Typeflag: tar.TypeReg}}, []string{"a", "b"})
	if _, err := readArchive(path); err == nil {
		t.Fatal("duplicate member accepted")
	}
}

func TestArchive_RegularTarAndZip(t *testing.T) {
	path := tarFixture(t, []*tar.Header{{Name: "man/", Mode: 0755, Typeflag: tar.TypeDir}, {Name: "man/tusk.1", Mode: 0644, Size: 2, Typeflag: tar.TypeReg}}, []string{"", "ok"})
	members, err := readArchive(path)
	if err != nil || len(members) != 1 || string(members["man/tusk.1"].Data) != "ok" {
		t.Fatalf("tar: %v", err)
	}
	path = filepath.Join(t.TempDir(), "fixture.zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(file)
	h := &zip.FileHeader{Name: "README.md", Method: zip.Deflate}
	h.SetMode(0644)
	w, err := zw.CreateHeader(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("ok")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	members, err = readArchive(path)
	if err != nil || len(members) != 1 || string(members["README.md"].Data) != "ok" {
		t.Fatalf("zip: %v", err)
	}
}

func TestMetadata_CompilerCGOAndPatchedGraph(t *testing.T) {
	good := func() *debug.BuildInfo {
		return &debug.BuildInfo{GoVersion: "go1.27.1", Main: debug.Module{Path: "github.com/newbpydev/tusk"},
			Settings: []debug.BuildSetting{{Key: "CGO_ENABLED", Value: "0"}, {Key: "GOOS", Value: "linux"}, {Key: "GOARCH", Value: "amd64"}, {Key: "-trimpath", Value: "true"}, {Key: "vcs.revision", Value: strings.Repeat("a", 40)}, {Key: "vcs.modified", Value: "false"}},
			Deps:     []*debug.Module{{Path: "github.com/charmbracelet/bubbletea", Version: "v1.3.10", Replace: &debug.Module{Path: "./third_party/bubbletea"}}, {Path: "github.com/charmbracelet/glamour", Version: "v0.9.1", Replace: &debug.Module{Path: "./third_party/glamour"}}}}
	}
	if err := validateBuildInfo(good(), "linux/amd64", "1.27.1", strings.Repeat("a", 40)); err != nil {
		t.Fatal(err)
	}
	if err := validateBuildInfo(good(), "linux", "1.27.1", strings.Repeat("a", 40)); err == nil || !strings.Contains(err.Error(), "invalid target") {
		t.Fatalf("malformed target accepted: %v", err)
	}
	for _, mutate := range []func(*debug.BuildInfo){
		func(v *debug.BuildInfo) { v.GoVersion = "go1.25.0" },
		func(v *debug.BuildInfo) { v.Settings[0].Value = "1" },
		func(v *debug.BuildInfo) { v.Settings[1].Value = "windows" },
		func(v *debug.BuildInfo) { v.Settings[3].Value = "false" },
		func(v *debug.BuildInfo) { v.Settings[4].Value = strings.Repeat("b", 40) },
		func(v *debug.BuildInfo) { v.Settings[5].Value = "true" },
		func(v *debug.BuildInfo) { v.Deps = v.Deps[:1] },
		func(v *debug.BuildInfo) { v.Deps[0].Replace = nil },
	} {
		v := good()
		mutate(v)
		if err := validateBuildInfo(v, "linux/amd64", "1.27.1", strings.Repeat("a", 40)); err == nil {
			t.Fatal("invalid build metadata accepted")
		}
	}
}
