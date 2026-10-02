package main

import (
	"bytes"
	"debug/buildinfo"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"runtime/debug"
	"strings"
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
	case "windows/amd64":
		file, err := pe.NewFile(bytes.NewReader(data))
		if err != nil {
			return err
		}
		defer file.Close()
		if file.Machine != pe.IMAGE_FILE_MACHINE_AMD64 {
			return fmt.Errorf("incorrect PE architecture")
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
	return nil
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
