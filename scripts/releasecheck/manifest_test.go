package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func command(t *testing.T, directory string, environment []string, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = directory
	cmd.Env = append(os.Environ(), environment...)
	data, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %q: %v\n%s", name, args, err, data)
	}
	return data
}

// Real cross-built Tusk executables exercise all three binary readers, embedded
// tzdata and the patched module graph. Cross-building is not native acceptance.
func releaseFixture(t *testing.T) (string, string, string, string) {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	checkout := filepath.Join(t.TempDir(), "source ü")
	if err := os.Mkdir(checkout, 0700); err != nil {
		t.Fatal(err)
	}
	paths := command(t, root, nil, "git", "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	for _, name := range strings.Split(strings.TrimSuffix(string(paths), "\x00"), "\x00") {
		// Keep the real app, local patches and payload. Historical evidence and
		// unrelated test suites add no inspector coverage and amplify race costs.
		if !strings.Contains(name, "/") {
			switch name {
			case "Makefile", ".goreleaser.yaml", "go.mod", "go.sum", "README.md", "LICENSE", "THIRD_PARTY_NOTICES.md", "CONTRIBUTING.md", "SECURITY.md":
			default:
				continue
			}
		} else if !strings.HasPrefix(name, "cmd/") && !strings.HasPrefix(name, "internal/") && !strings.HasPrefix(name, "db/") && !strings.HasPrefix(name, "third_party/") && !strings.HasPrefix(name, "docs/") && name != "scripts/tool-versions.json" && name != "scripts/notices.json" {
			continue
		}
		if strings.HasPrefix(name, "docs/") && !strings.HasPrefix(name, "docs/man/") && !strings.HasPrefix(name, "docs/completions/") && !strings.HasPrefix(name, "docs/assets/") && name != "docs/cli.md" && name != "docs/tui.md" && name != "docs/install.md" && name != "docs/service.md" {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		from := filepath.Join(root, name)
		st, err := os.Stat(from)
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(from)
		if err != nil {
			t.Fatal(err)
		}
		to := filepath.Join(checkout, name)
		if err := os.MkdirAll(filepath.Dir(to), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(to, data, st.Mode().Perm()); err != nil {
			t.Fatal(err)
		}
	}
	command(t, checkout, nil, "git", "init", "--quiet")
	command(t, checkout, nil, "git", "config", "tar.umask", "0022")
	command(t, checkout, nil, "git", "add", "--all")
	command(t, checkout, nil, "git", "-c", "user.name=Release fixture", "-c", "user.email=fixture@example.invalid", "commit", "--quiet", "-m", "isolated test fixture")
	sha := strings.TrimSpace(string(command(t, checkout, nil, "git", "rev-parse", "HEAD")))
	stage := t.TempDir()
	overlay := filepath.Join(stage, "overlay ü")
	if run([]string{"overlay", checkout, overlay, "v0.3.0", "candidate"}, io.Discard) != 0 {
		t.Fatal("fixture overlay")
	}
	dist := filepath.Join(stage, "dist")
	if err := os.Mkdir(dist, 0700); err != nil {
		t.Fatal(err)
	}
	payload, err := payloadFiles(checkout)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range targets() {
		parts := strings.Split(target, "/")
		filename := filepath.Join(stage, strings.ReplaceAll(target, "/", "_"))
		command(t, checkout, []string{"GOOS=" + parts[0], "GOARCH=" + parts[1], "TUSK_RELEASE_FIXTURE_BINARY=" + filename, "TUSK_RELEASE_OVERLAY=" + filepath.Join(overlay, "overlay.json")}, "make", "build-release-fixture")
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		payload[executableName(target)] = member{data, 0755}
		writeArchiveFixture(t, filepath.Join(dist, archiveName("0.3.0", target)), payload)
		delete(payload, executableName(target))
	}
	command(t, checkout, nil, "git", "archive", "--format=tar.gz", "--prefix=tusk-0.3.0/", "--output="+filepath.Join(dist, "tusk_0.3.0_source.tar.gz"), "HEAD")
	return checkout, dist, filepath.Join(overlay, "input.json"), sha
}

func writeArchiveFixture(t *testing.T, filename string, files map[string]member) {
	t.Helper()
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasSuffix(filename, ".zip") {
		out := zip.NewWriter(file)
		for name, member := range files {
			header := &zip.FileHeader{Name: name, Method: zip.Deflate}
			header.SetMode(member.Mode)
			w, err := out.CreateHeader(header)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write(member.Data); err != nil {
				t.Fatal(err)
			}
		}
		if err := out.Close(); err != nil {
			t.Fatal(err)
		}
	} else {
		gz := gzip.NewWriter(file)
		out := tar.NewWriter(gz)
		for name, member := range files {
			if err := out.WriteHeader(&tar.Header{Name: name, Mode: int64(member.Mode.Perm()), Size: int64(len(member.Data)), Typeflag: tar.TypeReg}); err != nil {
				t.Fatal(err)
			}
			if _, err := out.Write(member.Data); err != nil {
				t.Fatal(err)
			}
		}
		if err := out.Close(); err != nil {
			t.Fatal(err)
		}
		if err := gz.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestManifest_ActualFiveTargetRoundTripAndTamper(t *testing.T) {
	if testing.Short() {
		t.Skip("full cross-built archive integration; make test/test-release/race/validate covers it")
	}
	root, dist, input, sha := releaseFixture(t)
	assets := filepath.Join(t.TempDir(), "assets")
	if run([]string{"finalize", root, dist, assets, input, sha, "1.27.1", "candidate"}, os.Stderr) != 0 {
		t.Fatal("finalize")
	}
	if run([]string{"verify", assets}, os.Stderr) != 0 {
		t.Fatal("verify")
	}
	for _, target := range targets() {
		files, err := readArchive(filepath.Join(dist, archiveName("0.3.0", target)))
		if err != nil {
			t.Fatal(err)
		}
		data := append([]byte(nil), files[executableName(target)].Data...)
		switch target {
		case "darwin/amd64":
			binary.LittleEndian.PutUint32(data[4:], 0x100000c)
		case "darwin/arm64":
			binary.LittleEndian.PutUint32(data[4:], 0x1000007)
		case "windows/amd64":
			binary.LittleEndian.PutUint16(data[binary.LittleEndian.Uint32(data[60:])+4:], 0xaa64)
		default:
			continue
		}
		if inspectFormat(data, target) == nil {
			t.Fatal("wrong executable architecture")
		}
	}
	files, err := readArchive(filepath.Join(dist, archiveName("0.3.0", "linux/amd64")))
	if err != nil {
		t.Fatal(err)
	}
	if inspectBinary(bytes.ReplaceAll(files["tusk"].Data, []byte("time/tzdata.loadFromEmbeddedTZData"), []byte("time/tzdata.fakeFromEmbeddedTZData")), "linux/amd64", "1.27.1", sha) == nil {
		t.Fatal("missing tzdata accepted")
	}
	for _, change := range []func(map[string]member){
		func(m map[string]member) { delete(m, "tusk") },
		func(m map[string]member) { v := m["tusk"]; v.Mode = 0644; m["tusk"] = v },
		func(m map[string]member) { v := m["tusk"]; v.Data = []byte("invalid"); m["tusk"] = v },
		func(m map[string]member) { v := m["README.md"]; v.Data = []byte("stale"); m["README.md"] = v },
	} {
		bad := make(map[string]member, len(files))
		for k, v := range files {
			bad[k] = v
		}
		change(bad)
		path := filepath.Join(dist, archiveName("0.3.0", "linux/amd64"))
		writeArchiveFixture(t, path, bad)
		if finalize(root, dist, filepath.Join(t.TempDir(), "bad"), input, sha, "1.27.1", "candidate") == nil {
			t.Fatal("invalid payload finalized")
		}
		writeArchiveFixture(t, path, files)
	}
	if err := finalize(root, dist, assets, input, sha, "1.27.1", "candidate"); err == nil {
		t.Fatal("output collision accepted")
	}
	manifestPath := filepath.Join(assets, "release-manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	for index, mutate := range []func(*manifest){
		func(m *manifest) { m.Schema = 0 }, func(m *manifest) { m.SourceSHA = strings.Repeat("b", 40) },
		func(m *manifest) { m.Compiler = "1.25.0" }, func(m *manifest) { m.Version = "0.4.0" },
		func(m *manifest) { m.VersionInputSHA256 = "bad" }, func(m *manifest) { m.Targets = m.Targets[:4] },
		func(m *manifest) { m.Targets[0].Target = "darwin/amd64" }, func(m *manifest) { m.Targets[0].ExecutableSHA256 = "bad" },
		func(m *manifest) { m.Targets[0].Members["tusk"] = "bad" }, func(m *manifest) { m.Files["THIRD_PARTY_NOTICES.md"] = "bad" },
		func(m *manifest) { delete(m.Files, "THIRD_PARTY_NOTICES.md"); m.Files["release-manifest.json"] = "bad" },
		func(m *manifest) { m.ChecksumsSHA256 = "bad" }, func(m *manifest) { m.SourceMembers["tusk-0.3.0/LICENSE"] = "bad" },
		func(m *manifest) { m.Inputs[".goreleaser.yaml"] = "bad" }, func(m *manifest) { m.Mode = "published" },
		func(m *manifest) {
			delete(m.Inputs, "LICENSE")
			m.Inputs["README.md"] = m.SourceMembers["tusk-0.3.0/README.md"]
		},
	} {
		var m manifest
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatal(err)
		}
		mutate(&m)
		if err := writeJSON(manifestPath, m); err != nil {
			t.Fatal(err)
		}
		if err := verifyInventory(assets); err == nil {
			t.Fatalf("tampered manifest %d accepted", index)
		}
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	// A replacement symlink with identical bytes still cannot become an accepted asset.
	notice := filepath.Join(assets, "THIRD_PARTY_NOTICES.md")
	outside := filepath.Join(t.TempDir(), "notice")
	if err := os.Rename(notice, outside); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, notice); err != nil {
		if runtime.GOOS != "windows" {
			t.Fatal(err)
		}
		t.Logf("native Windows symlink privilege unavailable: %v", err)
	} else {
		if verifyInventory(assets) == nil {
			t.Fatal("symlink asset accepted")
		}
		if err := os.Remove(notice); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Rename(outside, notice); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{archiveName("0.3.0", "linux/amd64"), "tusk_0.3.0_source.tar.gz"} {
		file := filepath.Join(assets, name)
		temporary := filepath.Join(t.TempDir(), "retained")
		if err := os.Rename(file, temporary); err != nil {
			t.Fatal(err)
		}
		if verifyInventory(assets) == nil {
			t.Fatal("missing archive accepted")
		}
		if err := os.Rename(temporary, file); err != nil {
			t.Fatal(err)
		}
	}
	// Finalization fails before creating assets when a prerequisite/member is absent.
	for _, name := range []string{"LICENSE", "docs/cli.md", archiveName("0.3.0", "linux/amd64"), "tusk_0.3.0_source.tar.gz"} {
		base := root
		if strings.HasSuffix(name, ".gz") {
			base = dist
		}
		file := filepath.Join(base, name)
		temporary := filepath.Join(t.TempDir(), "retained")
		if err := os.Rename(file, temporary); err != nil {
			t.Fatal(err)
		}
		if finalize(root, dist, filepath.Join(t.TempDir(), "out"), input, sha, "1.27.1", "candidate") == nil {
			t.Fatal("missing finalization prerequisite")
		}
		if err := os.Rename(temporary, file); err != nil {
			t.Fatal(err)
		}
	}
	sourcePath := filepath.Join(dist, "tusk_0.3.0_source.tar.gz")
	source, err := readArchive(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	saved := source["tusk-0.3.0/LICENSE"]
	delete(source, "tusk-0.3.0/LICENSE")
	writeArchiveFixture(t, sourcePath, source)
	if finalize(root, dist, filepath.Join(t.TempDir(), "out"), input, sha, "1.27.1", "candidate") == nil {
		t.Fatal("incomplete source archive")
	}
	source["tusk-0.3.0/LICENSE"] = saved
	writeArchiveFixture(t, sourcePath, source)
	if err := os.WriteFile(filepath.Join(assets, "unexpected"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyInventory(assets); err == nil {
		t.Fatal("extra asset accepted")
	}
	if err := os.Remove(filepath.Join(assets, "unexpected")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assets, "checksums.txt"), []byte("corrupt"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := verifyInventory(assets); err == nil {
		t.Fatal("corrupt checksums accepted")
	}
}
