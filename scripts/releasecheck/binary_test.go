package main

import (
	"archive/zip"
	"bytes"
	"debug/pe"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
)

func peImportFixture() []byte {
	data := make([]byte, 512+4096)
	copy(data, "MZ")
	binary.LittleEndian.PutUint32(data[60:], 128)
	copy(data[128:], "PE\x00\x00")
	binary.LittleEndian.PutUint16(data[132:], pe.IMAGE_FILE_MACHINE_AMD64)
	binary.LittleEndian.PutUint16(data[134:], 1)
	binary.LittleEndian.PutUint16(data[148:], 240)
	header := data[152:392]
	binary.LittleEndian.PutUint16(header, 0x20b)
	binary.LittleEndian.PutUint32(header[108:], 16)
	binary.LittleEndian.PutUint32(header[120:], 0x1000)
	binary.LittleEndian.PutUint32(header[124:], 40)
	section := data[392:432]
	copy(section, ".rdata")
	binary.LittleEndian.PutUint32(section[8:], 4096)
	binary.LittleEndian.PutUint32(section[12:], 0x1000)
	binary.LittleEndian.PutUint32(section[16:], 4096)
	binary.LittleEndian.PutUint32(section[20:], 512)
	entry := data[512:532]
	binary.LittleEndian.PutUint32(entry, 0x1d00)
	binary.LittleEndian.PutUint32(entry[12:], 0x1c00)
	binary.LittleEndian.PutUint32(entry[16:], 0x1d00)
	copy(data[512+3072:], "KERNEL32.dll\x00")
	binary.LittleEndian.PutUint64(data[512+3328:], 1<<63|1) // ordinal-only import
	return data
}

func TestBinary_PEImportsOrdinalAndMalformedTables(t *testing.T) {
	file, err := pe.NewFile(bytes.NewReader(peImportFixture()))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	libraries, err := importedPELibraries(file)
	if err != nil || len(libraries) != 1 || libraries[0] != "KERNEL32.dll" {
		t.Fatalf("ordinal-only library omitted: %v %v", libraries, err)
	}
	if err := inspectFormat(peImportFixture(), "windows/amd64"); err != nil {
		t.Fatalf("system library positive control: %v", err)
	}
	for _, tc := range []struct {
		name, want string
		mutate     func(*pe.File, []byte)
	}{
		{"missing header", "missing PE import", func(f *pe.File, d []byte) { f.OptionalHeader = nil }},
		{"missing directory", "missing PE import", func(f *pe.File, d []byte) {
			f.OptionalHeader.(*pe.OptionalHeader64).NumberOfRvaAndSizes = 1
		}},
		{"outside section", "outside file", func(f *pe.File, d []byte) {
			f.OptionalHeader.(*pe.OptionalHeader64).DataDirectory[1].VirtualAddress = 0x5000
		}},
		{"virtual-only data", "outside file", func(f *pe.File, d []byte) {
			f.Sections[0].VirtualSize = 8192
			f.OptionalHeader.(*pe.OptionalHeader64).DataDirectory[1].VirtualAddress = 0x2000
		}},
		{"short section", "EOF", func(f *pe.File, d []byte) { f.Sections[0].Size = 8192 }},
		{"short directory", "table size", func(f *pe.File, d []byte) {
			f.OptionalHeader.(*pe.OptionalHeader64).DataDirectory[1].Size = 19
		}},
		{"oversized directory", "table size", func(f *pe.File, d []byte) {
			f.OptionalHeader.(*pe.OptionalHeader64).DataDirectory[1].Size = 8192
		}},
		{"invalid name address", "outside file", func(_ *pe.File, d []byte) { binary.LittleEndian.PutUint32(d[524:], 0) }},
		{"empty name", "library name", func(_ *pe.File, d []byte) { d[512+3072] = 0 }},
		{"unterminated name", "library name", func(_ *pe.File, d []byte) {
			for i := 512 + 3072; i < len(d); i++ {
				d[i] = 'x'
			}
		}},
		{"unterminated descriptors", "unterminated", func(f *pe.File, d []byte) {
			f.OptionalHeader.(*pe.OptionalHeader64).DataDirectory[1].Size = 20
		}},
		{"excessive descriptors", "excessive", func(f *pe.File, d []byte) {
			for i := 1; i < 128; i++ {
				copy(d[512+i*20:], d[512:532])
			}
			f.OptionalHeader.(*pe.OptionalHeader64).DataDirectory[1].Size = 128 * 20
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := peImportFixture()
			f, err := pe.NewFile(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			tc.mutate(f, data)
			if _, err := importedPELibraries(f); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("malformed import boundary: %v; want %s", err, tc.want)
			}
		})
	}
}

func TestBinary_PEPolicyAndBuildInfoBoundaries(t *testing.T) {
	data := peImportFixture()
	if err := inspectBinary(data, "windows/amd64", "1.27.1", strings.Repeat("a", 40)); err == nil || !strings.Contains(err.Error(), "not a Go executable") {
		t.Fatalf("format-valid binary without Go metadata: %v", err)
	}
	binary.LittleEndian.PutUint16(data[132:], pe.IMAGE_FILE_MACHINE_I386)
	if err := inspectFormat(data, "windows/amd64"); err == nil || !strings.Contains(err.Error(), "architecture") {
		t.Fatalf("wrong PE architecture: %v", err)
	}
	data = peImportFixture()
	binary.LittleEndian.PutUint32(data[272:], 0)
	if err := inspectFormat(data, "windows/amd64"); err == nil || !strings.Contains(err.Error(), "outside file") {
		t.Fatalf("invalid PE import table: %v", err)
	}
	for _, name := range []string{"ws2_32.dll", "advapi32.dll", "vendor.dll"} {
		data = peImportFixture()
		clear(data[512+3072 : 512+3328])
		copy(data[512+3072:], name)
		if err := inspectFormat(data, "windows/amd64"); err == nil || !strings.Contains(err.Error(), "unreviewed Windows library: "+name) {
			t.Fatalf("new import must request policy review: %v", err)
		}
	}
}

func timezoneZIPFixture(t *testing.T, count int, method uint16, size int) []byte {
	t.Helper()
	zone := make([]byte, size)
	if size >= 54 {
		copy(zone, "TZif")
		binary.BigEndian.PutUint32(zone[36:], 1)
		binary.BigEndian.PutUint32(zone[40:], 4)
		copy(zone[50:], "UTC\x00")
	}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("Fixture/%d", i)
		if i < 3 {
			name = []string{"America/New_York", "Asia/Tokyo", "Pacific/Auckland"}[i]
		}
		member, err := w.CreateHeader(&zip.FileHeader{Name: name, Method: method})
		if err != nil {
			t.Fatal(err)
		}
		data := zone
		if i >= 9 && len(zone) > 54 {
			data = zone[:54]
		}
		if _, err := member.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestBinary_TimezonePayloadMalformedAndBounded(t *testing.T) {
	good := timezoneZIPFixture(t, 100, zip.Store, 54)
	if err := inspectTimezonePayload(good); err != nil {
		t.Fatalf("populated ZIP control: %v", err)
	}
	badCentral := bytes.Clone(good)
	badCentral[bytes.Index(badCentral, []byte{'P', 'K', 1, 2})] = 'x'
	badLocal := bytes.Clone(good)
	badLocal[0] = 'x'
	badCRC := bytes.Clone(good)
	badCRC[bytes.Index(badCRC, []byte("TZif"))] = 'x'
	missingRegion := bytes.ReplaceAll(good, []byte("Asia/Tokyo"), []byte("Else/Where"))
	for _, data := range [][]byte{
		nil, {'P', 'K', 5, 6}, append([]byte{'P', 'K', 5, 6}, make([]byte, 18)...),
		badCentral, badLocal, badCRC, missingRegion,
		timezoneZIPFixture(t, 100, zip.Store, 0), timezoneZIPFixture(t, 99, zip.Store, 54), timezoneZIPFixture(t, 2001, zip.Store, 54),
		timezoneZIPFixture(t, 100, zip.Deflate, 54), timezoneZIPFixture(t, 100, zip.Store, (1<<20)+1),
	} {
		if err := inspectTimezonePayload(data); err == nil {
			t.Fatal("malformed timezone ZIP accepted")
		}
	}
}
