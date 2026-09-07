package core

import (
	"regexp"
	"sort"
	"strings"
)

type Tag string

var tagRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func NormalizeTag(raw string) (Tag, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "#")
	trimmed = strings.TrimSpace(trimmed)
	normalized := strings.ToLower(trimmed)

	if len(normalized) < 1 || len(normalized) > 32 {
		return "", ErrInvalidTag
	}

	if !tagRegex.MatchString(normalized) {
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

func (t Tag) String() string {
	return string(t)
}
