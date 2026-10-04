package filter

import (
	"errors"
	"regexp"
	"strings"
)

func parseTextMatch(value string) (condition, error) {
	leadingWildcard := strings.HasPrefix(value, "*")
	trailingWildcard := strings.HasSuffix(value, "*")
	text := strings.TrimPrefix(value, "*")
	text = strings.TrimSuffix(text, "*")

	if strings.Contains(text, "*") {
		return condition{}, errors.New("match only allows * at the beginning or end")
	}

	return condition{
		cost:    2,
		pattern: textPattern(text, leadingWildcard, trailingWildcard),
		match: func(line string) bool {
			switch {
			case leadingWildcard && trailingWildcard:
				return strings.Contains(line, text)
			case leadingWildcard:
				return strings.HasSuffix(line, text)
			case trailingWildcard:
				return strings.HasPrefix(line, text)
			default:
				return line == text
			}
		},
	}, nil
}

func textPattern(text string, leadingWildcard, trailingWildcard bool) string {
	pattern := regexp.QuoteMeta(text)
	if leadingWildcard {
		pattern = `(?s:.*)` + pattern
	}

	if trailingWildcard {
		pattern += `(?s:.*)`
	}

	return pattern
}
