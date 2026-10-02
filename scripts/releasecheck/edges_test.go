package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestArchive_CorruptionAndBounds(t *testing.T) {
	for _, content := range [][]byte{nil, []byte("bad gzip")} {
		file := filepath.Join(t.TempDir(), "bad.tar.gz")
		if err := os.WriteFile(file, content, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := readArchive(file); err == nil {
			t.Fatal("corrupt tar accepted")
		}
	}
	if _, err := readArchive(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing archive")
	}
	if _, err := readArchive(filepath.Join(t.TempDir(), "missing.zip")); err == nil {
		t.Fatal("missing zip")
	}
	file := tarFixture(t, []*tar.Header{{Name: "ok", Mode: 0644, Size: 1, Typeflag: tar.TypeReg}}, []string{"x"})
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-8] ^= 1
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readArchive(file); err == nil {
		t.Fatal("gzip footer corruption accepted")
	}
	file = filepath.Join(t.TempDir(), "unsafe.zip")
	out, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}
	z := zip.NewWriter(out)
	h := &zip.FileHeader{Name: "link"}
	h.SetMode(os.ModeSymlink | 0777)
	w, err := z.CreateHeader(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("target")); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := readArchive(file); err == nil {
		t.Fatal("zip symlink accepted")
	}
	for _, name := range []string{"", "a//b", "a/./b", ".env.local", "a.db-shm"} {
		if safeName(name) == nil {
			t.Fatalf("unsafe %q", name)
		}
	}
}

func TestArchive_RejectTrailingTarStream(t *testing.T) {
	file := tarFixture(t, []*tar.Header{{Name: "ok", Mode: 0644, Size: 1, Typeflag: tar.TypeReg}}, []string{"x"})
	out, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(out)
	if _, err := gz.Write([]byte("hidden unvalidated archive data")); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := readArchive(file); err == nil {
		t.Fatal("unvalidated trailing tar stream accepted")
	}
}

func TestArchive_GitGlobalMetadata(t *testing.T) {
	for _, records := range []map[string]string{{"comment": "invalid"}, {"comment": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "path": "../outside"}} {
		file := tarFixture(t, []*tar.Header{{Name: "pax_global_header", Typeflag: tar.TypeXGlobalHeader, PAXRecords: records}}, []string{""})
		if _, err := readArchive(file); err == nil {
			t.Fatal("unsafe global metadata accepted")
		}
	}
}

func TestInspector_BadFormatsAndELFInterpreter(t *testing.T) {
	for _, target := range append(targets(), "windows/arm64") {
		if inspectFormat([]byte("invalid"), target) == nil {
			t.Fatal("invalid format accepted")
		}
		if inspectBinary([]byte("invalid"), target, "1.27.1", "sha") == nil {
			t.Fatal("invalid binary accepted")
		}
	}
	// Minimal ELF64 with a PT_INTERP header, independent of native system tools.
	data := make([]byte, 120)
	copy(data, []byte{0x7f, 'E', 'L', 'F', 2, 1, 1})
	binary.LittleEndian.PutUint16(data[16:], 2)
	binary.LittleEndian.PutUint16(data[18:], 62)
	binary.LittleEndian.PutUint32(data[20:], 1)
	binary.LittleEndian.PutUint64(data[32:], 64)
	binary.LittleEndian.PutUint16(data[52:], 64)
	binary.LittleEndian.PutUint16(data[54:], 56)
	binary.LittleEndian.PutUint16(data[56:], 1)
	binary.LittleEndian.PutUint32(data[64:], 3)
	if inspectFormat(data, "linux/amd64") == nil {
		t.Fatal("dynamic interpreter accepted")
	}
	if inspectFormat(data, "linux/arm64") == nil {
		t.Fatal("wrong ELF machine accepted")
	}
}

func TestOverlay_SourceAndFilesystemFailures(t *testing.T) {
	root := t.TempDir()
	for _, source := range []string{filepath.Join(root, "missing"), root} {
		if _, err := createOverlay(source, filepath.Join(t.TempDir(), "out"), "v0.3.0", false); err == nil {
			t.Fatal("missing source accepted")
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "cmd", "tusk"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "cmd", "tusk", "version.go"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := createOverlay(root, filepath.Join(t.TempDir(), "out"), "v0.3.0", false); err == nil {
		t.Fatal("directory source accepted")
	}
	if _, err := createOverlay(root, filepath.Join(t.TempDir(), "missing", "out"), "v0.3.0", false); err == nil {
		t.Fatal("missing parent accepted")
	}
}

func TestFilesystem_ReadFailures(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("native ACL counterpart covers Windows read refusal")
	}
	root := t.TempDir()
	file := filepath.Join(root, "README.md")
	if err := os.WriteFile(file, []byte("x"), 0000); err != nil {
		t.Fatal(err)
	}
	if _, err := payloadFiles(root); err == nil {
		t.Fatal("unreadable payload")
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(file, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := payloadFiles(root); err == nil {
		t.Fatal("directory payload")
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	command(t, root, nil, "git", "init", "--quiet")
	command(t, root, nil, "git", "add", "README.md")
	if err := os.Chmod(file, 0000); err != nil {
		t.Fatal(err)
	}
	if _, err := sourceFiles(root, "p/"); err == nil {
		t.Fatal("unreadable source")
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if _, err := sourceFiles(root, "p/"); err == nil {
		t.Fatal("missing source")
	}
	command(t, root, nil, "git", "rm", "--cached", "README.md")
	if err := os.WriteFile(filepath.Join(root, "private.db"), []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	command(t, root, nil, "git", "add", "private.db")
	if _, err := sourceFiles(root, "p/"); err == nil {
		t.Fatal("private source data")
	}
}

func TestSource_TrackedExecutableMode(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "script.sh"), []byte("#!/bin/sh\n"), 0644); err != nil {
		t.Fatal(err)
	}
	command(t, root, nil, "git", "init", "--quiet")
	command(t, root, nil, "git", "add", "script.sh")
	command(t, root, nil, "git", "update-index", "--chmod=+x", "script.sh")
	files, err := sourceFiles(root, "source/")
	if err != nil || files["source/script.sh"].Mode.Perm() != 0755 {
		t.Fatalf("Git executable mode lost: %v %v", files, err)
	}
	command(t, root, nil, "git", "update-index", "--force-remove", "script.sh")
	if _, err := sourceFiles(root, "source/"); err == nil {
		t.Fatal("empty source index accepted")
	}
}

func TestOverlay_MissingWorkingDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows refuses removal of an active working directory")
	}
	root := t.TempDir()
	t.Chdir(t.TempDir())
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(cwd); err != nil {
		t.Fatal(err)
	}
	if _, err := createOverlay("relative", "out", "v0.3.0", false); err == nil {
		t.Fatal("missing working directory source")
	}
	if _, err := createOverlay(root, "out", "v0.3.0", false); err == nil {
		t.Fatal("missing working directory output")
	}
}

func TestManifest_ComponentsAndFailures(t *testing.T) {
	root := t.TempDir()
	if _, err := payloadFiles(root); err == nil {
		t.Fatal("empty payload inputs accepted")
	}
	if _, err := sourceFiles(root, "prefix/"); err == nil {
		t.Fatal("non-checkout source accepted")
	}
	good := map[string]member{"file": {[]byte("ok"), 0644}}
	if err := matchMembers(good, good); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []map[string]member{nil, {"file": {[]byte("wrong"), 0644}}, {"file": {[]byte("ok"), 0755}}, {"other": {[]byte("ok"), 0644}}} {
		if matchMembers(bad, good) == nil {
			t.Fatal("invalid members accepted")
		}
	}
	if matchHashes(good, nil) == nil {
		t.Fatal("missing hashes")
	}
	if matchHashes(good, map[string]string{"file": "bad"}) == nil {
		t.Fatal("wrong hash")
	}
	if err := matchHashes(good, memberHashes(good)); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(checksumText(map[string]string{"b": "2", "a": "1"}), []byte("1  a\n2  b\n")) {
		t.Fatal("checksum order")
	}
	if _, err := fileHash(root); err == nil {
		t.Fatal("directory hash")
	}
	if err := writeJSON(filepath.Join(root, "missing", "file"), map[string]string{}); err == nil {
		t.Fatal("missing output parent")
	}
	if err := writeJSON(filepath.Join(root, "file"), make(chan int)); err == nil {
		t.Fatal("invalid JSON value")
	}
	if err := readJSON(filepath.Join(root, "missing"), new(manifest)); err == nil {
		t.Fatal("missing input")
	}
	if err := os.WriteFile(filepath.Join(root, "bad"), []byte(`{"unknown":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := readJSON(filepath.Join(root, "bad"), new(manifest)); err == nil {
		t.Fatal("unknown input field")
	}
	if run([]string{"verify", root}, io.Discard) == 0 {
		t.Fatal("missing manifest")
	}
	if run([]string{"finalize", root, root, root, filepath.Join(root, "missing"), "sha", "1.27.1", "candidate"}, io.Discard) == 0 {
		t.Fatal("missing input receipt")
	}
	input := filepath.Join(root, "input.json")
	if err := writeJSON(input, overlayInfo{Version: "0.3.0", InputSHA256: "hash"}); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"bad", "snapshot", "candidate"} {
		if finalize(root, root, root, input, "sha", "1.27.1", mode) == nil {
			t.Fatal("invalid finalize succeeded")
		}
	}
}
