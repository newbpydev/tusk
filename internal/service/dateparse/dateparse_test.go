package dateparse

import (
	"encoding/binary"
	"errors"
	"github.com/newbpydev/tusk/internal/ports"
	"testing"
	"time"
	_ "time/tzdata"
)

func TestTonight_RepeatedWallTime(t *testing.T) {
	// A test-owned TZif transition repeats 19:30-20:30 on 2010-01-01.
	data := make([]byte, 44)
	copy(data, "TZif")
	binary.BigEndian.PutUint32(data[32:36], 1)
	binary.BigEndian.PutUint32(data[36:40], 2)
	binary.BigEndian.PutUint32(data[40:44], 4)
	data = binary.BigEndian.AppendUint32(data, uint32(time.Date(2010, 1, 1, 20, 30, 0, 0, time.UTC).Unix()))
	data = append(data, 1)
	data = binary.BigEndian.AppendUint32(data, 0)
	data = append(data, 0, 0)
	data = binary.BigEndian.AppendUint32(data, uint32(1<<32-3600))
	data = append(data, 0, 2, 'A', 0, 'B', 0)
	z, err := time.LoadLocationFromTZData("repeated-evening", data)
	if err != nil {
		t.Fatal(err)
	}
	r := time.Date(2010, 1, 1, 12, 0, 0, 0, time.UTC)
	got, err := ParseDue("tonight", r, z)
	if err != nil || !got.Equal(time.Date(2010, 1, 1, 20, 0, 0, 0, time.UTC)) {
		t.Fatalf("%v %v", got, err)
	}
}

func ref() time.Time { return time.Date(2026, 1, 31, 22, 0, 0, 0, time.UTC) }
func zone(t *testing.T, name string) *time.Location {
	t.Helper()
	z, e := time.LoadLocation(name)
	if e != nil {
		t.Fatal(e)
	}
	return z
}
func TestParseDue_Calendar(t *testing.T) {
	for _, tc := range []struct{ input, want string }{{"today", "2026-01-31"}, {"tomorrow", "2026-02-01"}, {"+1d", "2026-02-01"}, {"+3d", "2026-02-03"}, {"+1w", "2026-02-07"}, {"+2w", "2026-02-14"}, {"+1m", "2026-02-28"}, {"+2m", "2026-03-31"}, {" +001M ", "2026-02-28"}, {"mon", "2026-02-02"}, {"TUE", "2026-02-03"}, {"wed", "2026-02-04"}, {"thu", "2026-02-05"}, {"fri", "2026-02-06"}, {"sat", "2026-01-31"}, {"sun", "2026-02-01"}, {"2024-02-29", "2024-02-29"}} {
		t.Run(tc.input, func(t *testing.T) {
			v, e := ParseDue(tc.input, ref(), time.UTC)
			if e != nil || v.Format("2006-01-02") != tc.want || v.Hour() != 23 || v.Minute() != 59 || v.Second() != 59 || v.Nanosecond() != 999999999 || v.Location() != time.UTC {
				t.Fatalf("%v %v", v, e)
			}
		})
	}
	v, e := ParseDue("ToNight", ref(), time.UTC)
	if e != nil || v.Hour() != 20 || v.Day() != 31 {
		t.Fatalf("%v %v", v, e)
	}
	v, e = ParseDue("+1m", time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC), time.UTC)
	if e != nil || v.Day() != 29 {
		t.Fatalf("%v %v", v, e)
	}
	for _, s := range []string{"2026-02-01T01:02:03Z", "2026-02-01T01:02:03.1Z", "2026-02-01T01:02:03.123456789-03:00", "2026-02-01T01:02:03+23:59"} {
		v, e := ParseDue(s, ref(), time.UTC)
		want, _ := time.Parse(time.RFC3339Nano, s)
		if e != nil || !v.Equal(want) {
			t.Fatalf("%s: %v %v", s, v, e)
		}
	}
}
func TestParseDue_InvalidAndAmbiguous(t *testing.T) {
	for _, s := range []string{"", " ", "monday", "next week", "+0d", "+-1d", "+1.2d", "+1 d", "+1x", "+9999999999999999999999999w", "+99999999m", "+99999999d", "+99999999w", "2026-02-29", "0000-01-01", "2026-00-01", "2026-1-01", "2026-01-00", "2026-01-32", "2026-13-01", "2026-01-01t00:00:00Z", "2026-01-01T00:00:00z", "2026-01-01T00:00:60Z", "2026-01-01T00:00:00,1Z", "2026-01-01T00:00:00.1234567890Z", "2026-01-01T00:00:00+24:00", "2026-01-01T00:00:00+00:60", "2026-01-01T00:00:00", "2026-01-01T00:00:00Zx", "0001-01-01T00:00:00+23:59", "9999-12-31T23:59:59-23:59"} {
		t.Run(s, func(t *testing.T) {
			v, e := ParseDue(s, ref(), time.UTC)
			if !errors.Is(e, ports.ErrInvalidDate) || !v.IsZero() {
				t.Fatalf("accepted %q: %v %v", s, v, e)
			}
		})
	}
	for _, tc := range []struct{ s, z string }{{"2011-12-30", "Pacific/Apia"}, {"2018-11-04", "America/Sao_Paulo"}} {
		_, _, e := DayBounds(tc.s, ref(), zone(t, tc.z))
		if !errors.Is(e, ports.ErrInvalidDate) {
			t.Fatalf("%+v: %v", tc, e)
		}
	}
	z := zone(t, "America/New_York")
	v, e := wallTime(2026, 11, 1, 1, 30, 0, 0, z)
	if e != nil || v.Format(time.RFC3339) != "2026-11-01T05:30:00Z" {
		t.Fatalf("ambiguous: %v %v", v, e)
	}
	if _, e := wallTime(2026, 3, 8, 2, 30, 0, 0, z); !errors.Is(e, ports.ErrInvalidDate) {
		t.Fatal(e)
	}
}
func TestDayBounds_DST(t *testing.T) {
	z := zone(t, "America/New_York")
	for _, tc := range []struct {
		s     string
		hours int
	}{{"2026-03-08", 23}, {"2026-11-01", 25}} {
		a, b, e := DayBounds(tc.s, ref(), z)
		if e != nil || b.Sub(a) != time.Duration(tc.hours)*time.Hour || a.In(z).Hour() != 0 || b.In(z).Hour() != 0 {
			t.Fatalf("%v %v %v", a, b, e)
		}
	}
	a, b, e := DayBounds("2026-01-01T01:00:00+03:00", ref(), z)
	if e != nil || a.In(z).Format("2006-01-02") != "2025-12-31" || b.Sub(a) != 24*time.Hour {
		t.Fatalf("%v %v %v", a, b, e)
	}
	for _, s := range []string{"invalid", "9999-12-31", "2018-11-03"} {
		z := time.UTC
		if s == "2018-11-03" {
			z = zone(t, "America/Sao_Paulo")
		}
		a, b, e := DayBounds(s, ref(), z)
		if e == nil || !a.IsZero() || !b.IsZero() {
			t.Fatalf("partial bounds %s: %v %v %v", s, a, b, e)
		}
	}
}
func TestDateOptionsAndRange(t *testing.T) {
	for _, r := range []time.Time{{}, time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)} {
		if _, e := ParseDue("today", r, time.UTC); !errors.Is(e, ports.ErrInvalidReferenceTime) {
			t.Fatal(e)
		}
	}
	if _, e := ParseDue("today", ref(), nil); !errors.Is(e, ports.ErrInvalidDate) {
		t.Fatal(e)
	}
	if _, e := ParseDue("tomorrow", time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC), time.UTC); !errors.Is(e, ports.ErrInvalidDate) {
		t.Fatal(e)
	}
	v, e := ParseDue("today", ref(), zone(t, "America/Sao_Paulo"))
	if e != nil || v.Format(time.RFC3339Nano) != "2026-02-01T02:59:59.999999999Z" {
		t.Fatalf("%v %v", v, e)
	}
}

func TestDayBounds_FirstRepresentableDay(t *testing.T) {
	a, b, e := DayBounds("0001-01-01", ref(), time.UTC)
	if e != nil || !a.Equal(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)) || !b.Equal(time.Date(1, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("first day: %v %v %v", a, b, e)
	}
}
