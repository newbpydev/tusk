// Package dateparse implements the bounded task calendar grammar without I/O.
package dateparse

import (
	"github.com/newbpydev/tusk/internal/ports"
	"strconv"
	"strings"
	"time"
)

func valid(t time.Time) bool { return t.UTC().Year() >= 1 && t.UTC().Year() <= 9999 }

// ParseDue returns a UTC instant using only the supplied reference and location.
func ParseDue(expression string, reference time.Time, location *time.Location) (time.Time, error) {
	if reference.IsZero() || !valid(reference) {
		return time.Time{}, ports.ErrInvalidReferenceTime
	}
	if location == nil {
		return time.Time{}, ports.ErrInvalidDate
	}
	s := strings.TrimSpace(expression)
	if s == "" {
		return time.Time{}, ports.ErrInvalidDate
	}
	if len(s) > 10 && s[10] == 'T' {
		return timestamp(s)
	}
	local := reference.In(location)
	y, m, d := local.Date()
	hour, min, sec, ns := 23, 59, 59, 999999999
	day := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	word := strings.ToLower(s)
	switch word {
	case "today":
	case "tomorrow":
		day = day.AddDate(0, 0, 1)
	case "tonight":
		hour, min, sec, ns = 20, 0, 0, 0
	default:
		weekdays := []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}
		matched := false
		for i, w := range weekdays {
			if word == w {
				day = day.AddDate(0, 0, (i-int(local.Weekday())+7)%7)
				matched = true
				break
			}
		}
		if !matched {
			if word[0] == '+' {
				if len(word) < 3 {
					return time.Time{}, ports.ErrInvalidDate
				}
				digits := word[1 : len(word)-1]
				if !decimal(digits) {
					return time.Time{}, ports.ErrInvalidDate
				}
				n, err := strconv.ParseUint(digits, 10, 32)
				if err != nil || n == 0 {
					return time.Time{}, ports.ErrInvalidDate
				}
				switch word[len(word)-1] {
				case 'w':
					if n > 3652059/7 {
						return time.Time{}, ports.ErrInvalidDate
					}
					day = day.AddDate(0, 0, int(n)*7)
				case 'd':
					if n > 3652059 {
						return time.Time{}, ports.ErrInvalidDate
					}
					day = day.AddDate(0, 0, int(n))
				case 'm':
					if n > 119988 {
						return time.Time{}, ports.ErrInvalidDate
					}
					first := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC).AddDate(0, int(n), 0)
					last := first.AddDate(0, 1, -1).Day()
					if d > last {
						d = last
					}
					day = time.Date(first.Year(), first.Month(), d, 0, 0, 0, 0, time.UTC)
				default:
					return time.Time{}, ports.ErrInvalidDate
				}
			} else {
				var err error
				day, err = time.Parse("2006-01-02", s)
				if err != nil || len(s) != 10 {
					return time.Time{}, ports.ErrInvalidDate
				}
			}
		}
	}
	if !valid(day) {
		return time.Time{}, ports.ErrInvalidDate
	}
	y, m, d = day.Date()
	return wallTime(y, m, d, hour, min, sec, ns, location)
}

// DayBounds returns a half-open local day, rejecting missing midnight boundaries.
func DayBounds(expression string, reference time.Time, location *time.Location) (time.Time, time.Time, error) {
	due, err := ParseDue(expression, reference, location)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	y, m, d := due.In(location).Date()
	start, err := wallTime(y, m, d, 0, 0, 0, 0, location)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	next := time.Date(y, m, d+1, 0, 0, 0, 0, time.UTC)
	y, m, d = next.Date()
	end, err := wallTime(y, m, d, 0, 0, 0, 0, location)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}
func decimal(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range []byte(s) {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
func timestamp(s string) (time.Time, error) {
	if len(s) < 20 || s[10] != 'T' || s[13] != ':' || s[16] != ':' {
		return time.Time{}, ports.ErrInvalidDate
	}
	rest := s[19:]
	if rest[0] == '.' {
		i := 1
		for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
			i++
		}
		if i == 1 || i > 10 {
			return time.Time{}, ports.ErrInvalidDate
		}
		rest = rest[i:]
	}
	if rest != "Z" {
		if len(rest) != 6 || (rest[0] != '+' && rest[0] != '-') || rest[3] != ':' || !decimal(rest[1:3]) || !decimal(rest[4:]) || rest[1:3] > "23" || rest[4:] > "59" {
			return time.Time{}, ports.ErrInvalidDate
		}
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil || !valid(t) {
		return time.Time{}, ports.ErrInvalidDate
	}
	return t.UTC(), nil
}

// wallTime enumerates zone intervals surrounding a civil time, then round-trips
// candidates. This avoids time.Date's unspecified choice during repeated times.
func wallTime(y int, m time.Month, d, h, min, sec, ns int, loc *time.Location) (time.Time, error) {
	wall := time.Date(y, m, d, h, min, sec, ns, time.UTC)
	if !valid(wall) {
		return time.Time{}, ports.ErrInvalidDate
	}
	guess := time.Date(y, m, d, h, min, sec, ns, loc)
	offsets := make(map[int]bool)
	// Include the guess's interval for fixed zones, then adjacent transitions on
	// either side of both representations (also covers whole-date IANA skips).
	for _, anchor := range []time.Time{wall, guess} {
		cursor := anchor.Add(-48 * time.Hour)
		limit := anchor.Add(48 * time.Hour)
		for {
			at := cursor.In(loc)
			_, offset := at.Zone()
			offsets[offset] = true
			_, end := at.ZoneBounds()
			if end.IsZero() || end.After(limit) || !end.After(cursor) {
				break
			}
			cursor = end
		}
	}
	var earliest time.Time
	found := false
	for offset := range offsets {
		candidate := wall.Add(-time.Duration(offset) * time.Second)
		local := candidate.In(loc)
		if local.Year() == y && local.Month() == m && local.Day() == d && local.Hour() == h && local.Minute() == min && local.Second() == sec && local.Nanosecond() == ns && valid(candidate) {
			if !found || candidate.Before(earliest) {
				earliest = candidate
				found = true
			}
		}
	}
	if !found {
		return time.Time{}, ports.ErrInvalidDate
	}
	return earliest.UTC(), nil
}
