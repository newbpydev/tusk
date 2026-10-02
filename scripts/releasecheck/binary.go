package main

import (
	"archive/zip"
	"bytes"
	"debug/buildinfo"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"encoding/binary"
	"fmt"
	"io"
	"runtime/debug"
	"strings"
	"time"
)

func inspectFormat(data []byte, target string) error {
	switch target {
	case "linux/amd64", "linux/arm64":
		file, err := elf.NewFile(bytes.NewReader(data))
		if err != nil {
			return err
		}
		defer file.Close()
		want := elf.EM_X86_64
		if target == "linux/arm64" {
			want = elf.EM_AARCH64
		}
		if file.Machine != want || file.Class != elf.ELFCLASS64 {
			return fmt.Errorf("incorrect ELF architecture")
		}
		for _, program := range file.Progs {
			if program.Type == elf.PT_INTERP {
				return fmt.Errorf("Linux interpreter dependency")
			}
		}
		libraries, err := file.DynString(elf.DT_NEEDED)
		if err != nil || len(libraries) != 0 {
			return fmt.Errorf("Linux shared-library dependency: %v", err)
		}
	case "darwin/amd64", "darwin/arm64":
		file, err := macho.NewFile(bytes.NewReader(data))
		if err != nil {
			return err
		}
		defer file.Close()
		want := macho.CpuAmd64
		if target == "darwin/arm64" {
			want = macho.CpuArm64
		}
		if file.Cpu != want {
			return fmt.Errorf("incorrect Mach-O architecture")
		}
		libraries, err := file.ImportedLibraries()
		if err != nil {
			return err
		}
		for _, library := range libraries {
			if library != "/usr/lib/libSystem.B.dylib" && library != "/usr/lib/libresolv.9.dylib" {
				return fmt.Errorf("non-system macOS library: %s", library)
			}
		}
	case "windows/amd64":
		file, err := pe.NewFile(bytes.NewReader(data))
		if err != nil {
			return err
		}
		defer file.Close()
		if file.Machine != pe.IMAGE_FILE_MACHINE_AMD64 {
			return fmt.Errorf("incorrect PE architecture")
		}
		libraries, err := importedPELibraries(file)
		if err != nil {
			return err
		}
		for _, library := range libraries {
			// Lock imports to the pinned release's audited set. Even OS-shipped
			// additions require review before changing this allowlist.
			if !strings.EqualFold(library, "kernel32.dll") {
				return fmt.Errorf("unreviewed Windows library: %s", library)
			}
		}
	default:
		return fmt.Errorf("unsupported target %q", target)
	}
	return nil
}

func inspectBinary(data []byte, target, compiler, sha string) error {
	if err := inspectFormat(data, target); err != nil {
		return err
	}
	info, err := buildinfo.Read(bytes.NewReader(data))
	if err != nil {
		return err
	}
	if err := validateBuildInfo(info, target, compiler, sha); err != nil {
		return err
	}
	if !bytes.Contains(data, []byte("time/tzdata.loadFromEmbeddedTZData")) {
		return fmt.Errorf("missing embedded timezone data")
	}
	return inspectTimezonePayload(data)
}

// debug/pe.ImportedLibraries is unimplemented, and ImportedSymbols omits
// ordinal imports. Read every descriptor's DLL name, including ordinal-only
// entries, through bounded section-relative virtual addresses.
func importedPELibraries(file *pe.File) ([]string, error) {
	header, ok := file.OptionalHeader.(*pe.OptionalHeader64)
	if !ok || header.NumberOfRvaAndSizes <= pe.IMAGE_DIRECTORY_ENTRY_IMPORT {
		return nil, fmt.Errorf("missing PE import directory")
	}
	at := func(address uint32) ([]byte, error) {
		for _, section := range file.Sections {
			if address >= section.VirtualAddress && address-section.VirtualAddress < section.VirtualSize {
				data, err := section.Data()
				if err != nil {
					return nil, err
				}
				offset := uint64(address - section.VirtualAddress)
				if offset >= uint64(len(data)) {
					break
				}
				return data[offset:], nil
			}
		}
		return nil, fmt.Errorf("PE import address outside file")
	}
	directory := header.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_IMPORT]
	table, err := at(directory.VirtualAddress)
	if err != nil {
		return nil, err
	}
	if directory.Size < 20 || uint64(directory.Size) > uint64(len(table)) {
		return nil, fmt.Errorf("invalid PE import table size")
	}
	table = table[:directory.Size]
	var libraries []string
	for len(table) >= 20 && len(libraries) < 128 {
		entry := table[:20]
		table = table[20:]
		if bytes.Equal(entry, make([]byte, 20)) {
			return libraries, nil
		}
		name, err := at(binary.LittleEndian.Uint32(entry[12:16]))
		if err != nil {
			return nil, err
		}
		end := bytes.IndexByte(name, 0)
		if end <= 0 || end > 260 {
			return nil, fmt.Errorf("invalid PE library name")
		}
		libraries = append(libraries, string(name[:end]))
	}
	return nil, fmt.Errorf("unterminated or excessive PE import table")
}

// The standard library's embedded IANA data is a complete ZIP in rodata. A
// registration symbol alone also survives an empty/corrupt payload.
func inspectTimezonePayload(data []byte) error {
	for offset := 0; offset < len(data); {
		index := bytes.Index(data[offset:], []byte{'P', 'K', 5, 6})
		if index < 0 {
			break
		}
		endRecord := offset + index
		offset = endRecord + 4
		if endRecord+22 > len(data) {
			continue
		}
		record := data[endRecord : endRecord+22]
		start := int64(endRecord) - int64(binary.LittleEndian.Uint32(record[12:16])) - int64(binary.LittleEndian.Uint32(record[16:20]))
		end := endRecord + 22 + int(binary.LittleEndian.Uint16(record[20:22]))
		if start < 0 || start >= int64(endRecord) || end > len(data) {
			continue
		}
		archive, err := zip.NewReader(bytes.NewReader(data[start:end]), int64(end)-start)
		if err != nil || len(archive.File) < 100 || len(archive.File) > 2000 {
			continue
		}
		regions := map[string]bool{"America/New_York": false, "Asia/Tokyo": false, "Pacific/Auckland": false}
		valid := true
		var total uint64
		for _, file := range archive.File {
			total += file.UncompressedSize64
			if file.Method != zip.Store || file.UncompressedSize64 > 1<<20 || total > 8<<20 {
				valid = false
				break
			}
			reader, err := file.Open()
			if err != nil {
				valid = false
				break
			}
			zone, readErr := io.ReadAll(io.LimitReader(reader, 1<<20+1))
			closeErr := reader.Close()
			if readErr != nil || closeErr != nil || len(zone) > 1<<20 {
				valid = false
				break
			}
			if _, err := time.LoadLocationFromTZData(file.Name, zone); err != nil {
				valid = false
				break
			}
			if _, ok := regions[file.Name]; ok {
				regions[file.Name] = true
			}
		}
		if valid && regions["America/New_York"] && regions["Asia/Tokyo"] && regions["Pacific/Auckland"] {
			return nil
		}
	}
	return fmt.Errorf("missing or corrupt populated embedded timezone data")
}

func validateBuildInfo(info *debug.BuildInfo, target, compiler, sha string) error {
	if info.GoVersion != "go"+compiler || info.Main.Path != "github.com/newbpydev/tusk" {
		return fmt.Errorf("incorrect compiler or module")
	}
	parts := strings.Split(target, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid target %q", target)
	}
	settings := make(map[string]string)
	for _, s := range info.Settings {
		settings[s.Key] = s.Value
	}
	for key, value := range map[string]string{"CGO_ENABLED": "0", "GOOS": parts[0], "GOARCH": parts[1], "-trimpath": "true", "vcs.revision": sha, "vcs.modified": "false"} {
		if settings[key] != value {
			return fmt.Errorf("build setting %q: expected %q, got %q", key, value, settings[key])
		}
	}
	for name, version := range map[string]string{"bubbletea": "v1.3.10", "glamour": "v0.9.1"} {
		found := false
		for _, dep := range info.Deps {
			if dep.Path == "github.com/charmbracelet/"+name && dep.Version == version && dep.Replace != nil && dep.Replace.Path == "./third_party/"+name {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("missing pinned local replacement %q", name)
		}
	}
	return nil
}
