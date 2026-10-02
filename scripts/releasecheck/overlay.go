package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type overlayInfo struct {
	Version     string `json:"version"`
	InputSHA256 string `json:"input_sha256"`
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func createOverlay(root, destination, tag string, snapshot bool) (overlayInfo, error) {
	info := overlayInfo{}
	if !regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`).MatchString(tag) {
		return info, fmt.Errorf("invalid release version %q", tag)
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return info, err
	}
	destination, err = filepath.Abs(destination)
	if err != nil {
		return info, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return info, err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(destination))
	if err != nil {
		return info, err
	}
	destination = filepath.Join(parent, filepath.Base(destination))
	rel, err := filepath.Rel(root, destination)
	if err != nil || rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
		return info, fmt.Errorf("overlay must be outside the checkout")
	}
	source := filepath.Join(root, "cmd", "tusk", "version.go")
	st, err := os.Lstat(source)
	if err != nil || !st.Mode().IsRegular() {
		return info, fmt.Errorf("invalid version source")
	}
	if err := os.Mkdir(destination, 0700); err != nil {
		return info, err
	}
	info.Version = strings.TrimPrefix(tag, "v")
	if snapshot {
		info.Version += "-dev"
	}
	input := []byte(fmt.Sprintf("package main\n\nconst Version = %q\n", info.Version))
	info.InputSHA256 = digest(input)
	generated := filepath.Join(destination, "version.go")
	if err := os.WriteFile(generated, input, 0600); err != nil {
		return info, err
	}
	if err := writeJSON(filepath.Join(destination, "overlay.json"), struct{ Replace map[string]string }{map[string]string{source: generated}}); err != nil {
		return info, err
	}
	return info, nil
}
