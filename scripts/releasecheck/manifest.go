package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type targetRecord struct {
	Target           string            `json:"target"`
	Archive          string            `json:"archive"`
	ExecutableSHA256 string            `json:"executable_sha256"`
	Members          map[string]string `json:"members_sha256"`
}

func validateIdentity(m manifest) error {
	version := strings.TrimSuffix(m.Version, "-dev")
	if !regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`).MatchString(version) ||
		!regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(m.SourceSHA) ||
		(m.Mode != "candidate" && m.Mode != "snapshot") ||
		(m.Mode == "snapshot") != strings.HasSuffix(m.Version, "-dev") {
		return fmt.Errorf("invalid candidate identity")
	}
	return nil
}

type manifest struct {
	Schema             int               `json:"schema"`
	Version            string            `json:"version"`
	Mode               string            `json:"mode"`
	SourceSHA          string            `json:"source_sha"`
	Compiler           string            `json:"compiler"`
	VersionInputSHA256 string            `json:"version_input_sha256"`
	Inputs             map[string]string `json:"inputs_sha256"`
	Files              map[string]string `json:"files_sha256"`
	Targets            []targetRecord    `json:"targets"`
	SourceMembers      map[string]string `json:"source_members_sha256"`
	ChecksumsSHA256    string            `json:"checksums_sha256"`
}

func targets() []string {
	return []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64"}
}
func inputNames() []string {
	return []string{"scripts/tool-versions.json", ".goreleaser.yaml", "go.mod", "go.sum", "scripts/notices.json", "LICENSE", "THIRD_PARTY_NOTICES.md", "third_party/bubbletea/TUSK-PATCH.json", "third_party/glamour/TUSK-PATCH.json"}
}
func archiveName(version, target string) string {
	ext := ".tar.gz"
	if strings.HasPrefix(target, "windows/") {
		ext = ".zip"
	}
	return "tusk_" + version + "_" + strings.ReplaceAll(target, "/", "_") + ext
}
func executableName(target string) string {
	if strings.HasPrefix(target, "windows/") {
		return "tusk.exe"
	}
	return "tusk"
}
func writeJSON(filename string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, append(data, '\n'), 0644)
}
func readJSON(filename string, value any) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	return d.Decode(value)
}
func fileHash(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return digest(data), nil
}
func memberHashes(files map[string]member) map[string]string {
	hashes := make(map[string]string, len(files))
	for name, file := range files {
		hashes[name] = digest(file.Data)
	}
	return hashes
}

func payloadFiles(root string) (map[string]member, error) {
	files := make(map[string]member)
	for _, pattern := range []string{"README.md", "LICENSE", "THIRD_PARTY_NOTICES.md", "CONTRIBUTING.md", "SECURITY.md", "docs/completions/*", "docs/man/*.1", "docs/cli.md", "docs/tui.md", "docs/install.md", "docs/service.md", "docs/assets/*"} {
		paths, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(pattern)))
		if err != nil || len(paths) == 0 {
			return nil, fmt.Errorf("missing payload input %q", pattern)
		}
		for _, filename := range paths {
			st, err := os.Lstat(filename)
			if err != nil || !st.Mode().IsRegular() {
				return nil, fmt.Errorf("unsafe payload input %q", filename)
			}
			data, err := os.ReadFile(filename)
			if err != nil {
				return nil, err
			}
			rel, _ := filepath.Rel(root, filename)
			name := filepath.ToSlash(rel)
			if strings.HasPrefix(name, "docs/completions/") || strings.HasPrefix(name, "docs/man/") {
				name = strings.TrimPrefix(name, "docs/")
			}
			files[name] = member{data, 0644}
		}
	}
	return files, nil
}

func matchMembers(got, want map[string]member) error {
	if len(got) != len(want) {
		return fmt.Errorf("archive member count: got %d, want %d", len(got), len(want))
	}
	for name, expected := range want {
		actual, ok := got[name]
		if !ok || !bytes.Equal(actual.Data, expected.Data) || actual.Mode.Perm() != expected.Mode.Perm() {
			return fmt.Errorf("archive member mismatch %q", name)
		}
	}
	return nil
}

func sourceFiles(root, prefix string) (map[string]member, error) {
	cmd := exec.Command("git", "-C", root, "ls-files", "--stage", "-z")
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	files := make(map[string]member)
	for _, entry := range strings.Split(strings.TrimSuffix(string(data), "\x00"), "\x00") {
		metadata, name, ok := strings.Cut(entry, "\t")
		fields := strings.Fields(metadata)
		if !ok || len(fields) != 3 || fields[2] != "0" || (fields[0] != "100644" && fields[0] != "100755") {
			return nil, fmt.Errorf("unsupported Git source entry %q", entry)
		}
		if err := safeName(name); err != nil {
			return nil, err
		}
		filename := filepath.Join(root, filepath.FromSlash(name))
		st, err := os.Lstat(filename)
		if err != nil || !st.Mode().IsRegular() {
			return nil, fmt.Errorf("unsafe source input %q", name)
		}
		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		mode := os.FileMode(0644)
		if fields[0] == "100755" {
			mode = 0755
		}
		files[prefix+name] = member{content, mode}
	}
	return files, nil
}

func checksumText(files map[string]string) []byte {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var out strings.Builder
	for _, name := range names {
		fmt.Fprintf(&out, "%s  %s\n", files[name], name)
	}
	return []byte(out.String())
}

func finalize(root, dist, assets, input, sha, compiler, mode string) error {
	var version overlayInfo
	if err := readJSON(input, &version); err != nil {
		return err
	}
	if mode != "candidate" && mode != "snapshot" {
		return fmt.Errorf("invalid release mode")
	}
	if (mode == "snapshot") != strings.HasSuffix(version.Version, "-dev") {
		return fmt.Errorf("version/mode mismatch")
	}
	m := manifest{Schema: 1, Version: version.Version, Mode: mode, SourceSHA: sha, Compiler: compiler, VersionInputSHA256: version.InputSHA256, Inputs: make(map[string]string), Files: make(map[string]string)}
	if err := validateIdentity(m); err != nil {
		return err
	}
	for _, name := range inputNames() {
		var err error
		m.Inputs[name], err = fileHash(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
	}
	expected, err := payloadFiles(root)
	if err != nil {
		return err
	}
	// Validate the complete set before creating the destination or copying any bytes.
	for _, target := range targets() {
		name := archiveName(m.Version, target)
		files, err := readArchive(filepath.Join(dist, name))
		if err != nil {
			return err
		}
		binaryName := executableName(target)
		binary, ok := files[binaryName]
		if !ok || binary.Mode.Perm() != 0755 {
			return fmt.Errorf("missing executable/mode for %q", target)
		}
		if err := inspectBinary(binary.Data, target, compiler, sha); err != nil {
			return fmt.Errorf("%s: %w", target, err)
		}
		delete(files, binaryName)
		if err := matchMembers(files, expected); err != nil {
			return err
		}
		files[binaryName] = binary
		m.Targets = append(m.Targets, targetRecord{target, name, digest(binary.Data), memberHashes(files)})
	}
	sourceName := "tusk_" + m.Version + "_source.tar.gz"
	files, err := readArchive(filepath.Join(dist, sourceName))
	if err != nil {
		return err
	}
	want, err := sourceFiles(root, "tusk-"+m.Version+"/")
	if err != nil {
		return err
	}
	if err := matchMembers(files, want); err != nil {
		return err
	}
	m.SourceMembers = memberHashes(files)
	if err := os.Mkdir(assets, 0755); err != nil {
		return err
	}
	names := []string{sourceName, "THIRD_PARTY_NOTICES.md"}
	for _, target := range m.Targets {
		names = append(names, target.Archive)
	}
	for _, name := range names {
		from := filepath.Join(dist, name)
		if name == "THIRD_PARTY_NOTICES.md" {
			from = filepath.Join(root, name)
		}
		data, err := os.ReadFile(from)
		if err != nil {
			return err
		}
		m.Files[name] = digest(data)
		if err := os.WriteFile(filepath.Join(assets, name), data, 0644); err != nil {
			return err
		}
	}
	checksums := checksumText(m.Files)
	m.ChecksumsSHA256 = digest(checksums)
	if err := os.WriteFile(filepath.Join(assets, "checksums.txt"), checksums, 0644); err != nil {
		return err
	}
	return writeJSON(filepath.Join(assets, "release-manifest.json"), m)
}

func verifyInventory(assets string) error {
	var m manifest
	if err := readJSON(filepath.Join(assets, "release-manifest.json"), &m); err != nil {
		return err
	}
	if m.Schema != 1 || len(m.Targets) != 5 || len(m.Files) != 7 || len(m.SourceMembers) == 0 || len(m.Inputs) != 9 {
		return fmt.Errorf("incomplete manifest")
	}
	if err := validateIdentity(m); err != nil {
		return err
	}
	input := []byte(fmt.Sprintf("package main\n\nconst Version = %q\n", m.Version))
	if digest(input) != m.VersionInputSHA256 {
		return fmt.Errorf("incorrect generated input")
	}
	for index, target := range targets() {
		record := m.Targets[index]
		if record.Target != target || record.Archive != archiveName(m.Version, target) {
			return fmt.Errorf("incorrect target inventory")
		}
		files, err := readArchive(filepath.Join(assets, record.Archive))
		if err != nil {
			return err
		}
		if err := matchHashes(files, record.Members); err != nil {
			return err
		}
		binary := files[executableName(target)]
		if digest(binary.Data) != record.ExecutableSHA256 || binary.Mode.Perm() != 0755 {
			return fmt.Errorf("incorrect executable identity")
		}
		if err := inspectBinary(binary.Data, target, m.Compiler, m.SourceSHA); err != nil {
			return err
		}
		for name, file := range files {
			if name != executableName(target) && file.Mode.Perm() != 0644 {
				return fmt.Errorf("unexpected executable member %q", name)
			}
		}
	}
	sourceName := "tusk_" + m.Version + "_source.tar.gz"
	files, err := readArchive(filepath.Join(assets, sourceName))
	if err != nil {
		return err
	}
	if err := matchHashes(files, m.SourceMembers); err != nil {
		return err
	}
	prefix := "tusk-" + m.Version + "/"
	for _, name := range inputNames() {
		hash, ok := m.Inputs[name]
		if !ok || m.SourceMembers[prefix+name] != hash {
			return fmt.Errorf("source/build input mismatch %q", name)
		}
	}
	var pins struct {
		Go struct {
			Release string `json:"release"`
		} `json:"go"`
	}
	if err := json.Unmarshal(files[prefix+"scripts/tool-versions.json"].Data, &pins); err != nil || pins.Go.Release != m.Compiler {
		return fmt.Errorf("compiler does not match source tool lock")
	}
	expected := map[string]bool{sourceName: true, "THIRD_PARTY_NOTICES.md": true}
	for _, t := range m.Targets {
		expected[t.Archive] = true
	}
	for name, hash := range m.Files {
		if !expected[name] {
			return fmt.Errorf("unexpected asset %q", name)
		}
		actual, err := fileHash(filepath.Join(assets, name))
		if err != nil || hash != actual {
			return fmt.Errorf("asset digest mismatch %q", name)
		}
	}
	text, err := os.ReadFile(filepath.Join(assets, "checksums.txt"))
	if err != nil || digest(text) != m.ChecksumsSHA256 || !bytes.Equal(text, checksumText(m.Files)) {
		return fmt.Errorf("checksum mismatch")
	}
	entries, err := os.ReadDir(assets)
	if err != nil {
		return err
	}
	if len(entries) != 9 {
		return fmt.Errorf("unexpected asset count")
	}
	for _, entry := range entries {
		if entry.Type() != 0 || (!expected[entry.Name()] && entry.Name() != "checksums.txt" && entry.Name() != "release-manifest.json") {
			return fmt.Errorf("unexpected asset entry %q", entry.Name())
		}
	}
	return nil
}

func matchHashes(files map[string]member, hashes map[string]string) error {
	if len(files) != len(hashes) {
		return fmt.Errorf("member hash count mismatch")
	}
	for name, file := range files {
		if hashes[name] != digest(file.Data) {
			return fmt.Errorf("member hash mismatch %q", name)
		}
	}
	return nil
}
