package service

import (
	"encoding/hex"
	"github.com/newbpydev/tusk/internal/ports"
	"io"
	"time"
)

// NewUUIDv7 uses fresh caller-owned entropy; it makes no monotonic-order promise.
// A supplied reader must support Read; panics from unusable readers propagate.
func NewUUIDv7(now time.Time, entropy io.Reader) (string, error) {
	ms := now.UnixMilli()
	if !validTime(now) || ms < 0 || ms > 1<<48-1 {
		return "", ports.ErrInvalidReferenceTime
	}
	if entropy == nil {
		return "", ports.ErrIdentityGeneration
	}
	var b [16]byte
	if _, err := io.ReadFull(entropy, b[6:]); err != nil {
		return "", ports.ErrIdentityGeneration
	}
	for i := 5; i >= 0; i-- {
		b[i] = byte(ms)
		ms >>= 8
	}
	b[6] = b[6]&15 | 0x70
	b[8] = b[8]&63 | 0x80
	var out [36]byte
	hex.Encode(out[:8], b[:4])
	out[8] = '-'
	hex.Encode(out[9:13], b[4:6])
	out[13] = '-'
	hex.Encode(out[14:18], b[6:8])
	out[18] = '-'
	hex.Encode(out[19:23], b[8:10])
	out[23] = '-'
	hex.Encode(out[24:], b[10:])
	return string(out[:]), nil
}
func validUUIDv7(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range []byte(s) {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return s[14] == '7' && (s[19] == '8' || s[19] == '9' || s[19] == 'a' || s[19] == 'b')
}
