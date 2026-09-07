package core

import (
	"sort"
	"strings"
)

type Tag string

func isValidTag(s string) bool {
	if len(s) == 0 {
		return false
	}
	prevHyphen := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			prevHyphen = false
			continue
		}
		if c == '-' {
			if i == 0 || prevHyphen {
				return false
			}
			prevHyphen = true
			continue
		}
		return false
	}
	return !prevHyphen
}
func NormalizeTag(raw string) (Tag, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "#")
	trimmed = strings.TrimSpace(trimmed)
	normalized := strings.ToLower(trimmed)
	if len(normalized) < 1 || len(normalized) > 32 {
		return "", ErrInvalidTag
	}

	if !isValidTag(normalized) {
		return "", ErrInvalidTag
	}

	return Tag(normalized), nil
}

func NormalizeTags(raw []string) ([]Tag, error) {
	if len(raw) == 0 {
		return []Tag{}, nil
	}

	seen := make(map[Tag]struct{}, len(raw))
	result := make([]Tag, 0, len(raw))

	for _, s := range raw {
		tag, err := NormalizeTag(s)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[tag]; !exists {
			seen[tag] = struct{}{}
			result = append(result, tag)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result, nil
}

func NormalizeTagSlice(raw []Tag) ([]Tag, error) {
	if len(raw) == 0 {
		return []Tag{}, nil
	}
	converted := make([]string, len(raw))
	for i, t := range raw {
		converted[i] = string(t)
	}
	return NormalizeTags(converted)
}

func (t Tag) String() string {
	return string(t)
}
