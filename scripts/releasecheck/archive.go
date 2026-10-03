package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
	"unicode"
)

const memberLimit = 64 << 20
const archiveLimit = 512 << 20

type member struct {
	Data []byte
	Mode os.FileMode
}

func safeName(name string) error {
	if name == "" || strings.HasPrefix(name, "/") || strings.ContainsAny(name, "\\:") || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return fmt.Errorf("unsafe member %q", name)
	}
	for _, part := range strings.Split(strings.TrimSuffix(name, "/"), "/") {
		if part == "" || part == "." || part == ".." || part == ".git" || strings.HasPrefix(part, ".env") || strings.HasSuffix(part, ".db") || strings.HasSuffix(part, "-wal") || strings.HasSuffix(part, "-shm") {
			return fmt.Errorf("unsafe member %q", name)
		}
	}
	return nil
}

func readArchive(filename string) (map[string]member, error) {
	files := make(map[string]member)
	seen := make(map[string]bool)
	gitHeader := false
	var total int64
	add := func(name string, mode os.FileMode, size int64, reader io.Reader) error {
		if err := safeName(name); err != nil {
			return err
		}
		if mode&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky) != 0 {
			return fmt.Errorf("unsafe member permissions %q", name)
		}
		key := strings.TrimSuffix(name, "/")
		if seen[key] || len(seen) >= 10000 {
			return fmt.Errorf("duplicate or excessive members: %q", name)
		}
		seen[key] = true
		if mode.IsDir() {
			return nil
		}
		if !mode.IsRegular() || size < 0 || size > memberLimit || total+size > archiveLimit {
			return fmt.Errorf("unsafe type or size: %q", name)
		}
		data, err := io.ReadAll(io.LimitReader(reader, size+1))
		if err != nil || int64(len(data)) != size {
			return fmt.Errorf("invalid member data: %q: %v", name, err)
		}
		total += size
		files[path.Clean(name)] = member{data, mode}
		return nil
	}
	if strings.HasSuffix(filename, ".zip") {
		zr, err := zip.OpenReader(filename)
		if err != nil {
			return nil, err
		}
		defer zr.Close()
		for _, file := range zr.File {
			if file.UncompressedSize64 > memberLimit {
				return nil, fmt.Errorf("oversized member %q", file.Name)
			}
			r, err := file.Open()
			if err != nil {
				return nil, err
			}
			err = add(file.Name, file.Mode(), int64(file.UncompressedSize64), r)
			r.Close()
			if err != nil {
				return nil, err
			}
		}
		return files, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			// tar EOF precedes the gzip footer: finish reading to check its CRC.
			tail, err := io.ReadAll(io.LimitReader(gz, (1<<20)+1))
			if err != nil {
				return nil, err
			}
			if len(tail) > 1<<20 || strings.Trim(string(tail), "\x00") != "" {
				return nil, fmt.Errorf("unvalidated data after tar EOF")
			}
			return files, nil
		}
		if err != nil {
			return nil, err
		}
		// git archive emits a global PAX commit comment, not an extracted file.
		if h.Typeflag == tar.TypeXGlobalHeader {
			if gitHeader || len(seen) != 0 || h.Name != "pax_global_header" || len(h.PAXRecords) != 1 || !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(h.PAXRecords["comment"]) {
				return nil, fmt.Errorf("unsafe global archive metadata")
			}
			gitHeader = true
			continue
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA && h.Typeflag != tar.TypeDir {
			return nil, fmt.Errorf("unsafe member type %q", h.Name)
		}
		if err := add(h.Name, h.FileInfo().Mode(), h.Size, tr); err != nil {
			return nil, err
		}
	}
}
