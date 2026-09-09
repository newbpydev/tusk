package service

import (
	"bytes"
	"errors"
	"github.com/newbpydev/tusk/internal/ports"
	"io"
	"strings"
	"testing"
	"time"
)

type failingEntropy struct{}

func (failingEntropy) Read([]byte) (int, error) { return 0, errors.New("PRIVATE-ENTROPY") }
func TestUUIDv7_LayoutAndFailure(t *testing.T) {
	id, e := NewUUIDv7(time.UnixMilli(0x017f22e279b0), bytes.NewReader([]byte{0x0c, 0xc3, 0x18, 0xc4, 0xdc, 0x0c, 0x0c, 0x07, 0x39, 0x8f}))
	if e != nil || id != "017f22e2-79b0-7cc3-98c4-dc0c0c07398f" {
		t.Fatalf("%s %v", id, e)
	}
	for _, r := range []io.Reader{nil, bytes.NewReader(nil), bytes.NewReader(make([]byte, 9)), failingEntropy{}} {
		id, e := NewUUIDv7(time.UnixMilli(0), r)
		if id != "" || !errors.Is(e, ports.ErrIdentityGeneration) || strings.Contains(e.Error(), "PRIVATE") {
			t.Fatalf("%s %v", id, e)
		}
	}
	for _, r := range []time.Time{time.UnixMilli(-1), time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)} {
		if _, e := NewUUIDv7(r, failingEntropy{}); !errors.Is(e, ports.ErrInvalidReferenceTime) {
			t.Fatal(e)
		}
	}
	for _, s := range []string{"", strings.ToUpper(id), "017f22e2-79b0-4cc3-98c4-dc0c0c07398f", "017f22e2-79b0-7cc3-78c4-dc0c0c07398f", "g17f22e2-79b0-7cc3-98c4-dc0c0c07398f"} {
		if validUUIDv7(s) {
			t.Fatalf("accepted %q", s)
		}
	}
	if !validUUIDv7(id) {
		t.Fatal("valid rejected")
	}
}

func TestUUIDv7_IndependentEntropyAndClockReversal(t *testing.T) {
	for i := 0; i < 32; i++ {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			t.Parallel()
			now := time.UnixMilli(1000 - int64(i))
			raw := bytes.Repeat([]byte{byte(i)}, 10)
			a, e := NewUUIDv7(now, bytes.NewReader(raw))
			if e != nil || !validUUIDv7(a) {
				t.Fatalf("%s %v", a, e)
			}
			raw[9]++
			b, e := NewUUIDv7(now, bytes.NewReader(raw))
			if e != nil || a == b {
				t.Fatalf("entropy not used: %s %s %v", a, b, e)
			}
		})
	}
	if _, e := NewUUIDv7(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC), bytes.NewReader(make([]byte, 10))); e != nil {
		t.Fatal(e)
	}
	if validUUIDv7("017f22e2_79b0-7cc3-98c4-dc0c0c07398f") {
		t.Fatal("separator accepted")
	}
}
