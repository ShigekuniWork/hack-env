package filter

import (
	"fmt"
	"strings"
)

const maxRegexRepetition = 65535

// Regex generates a PCRE2 expression matching every condition in UTF mode.
// Multiple conditions are combined with lookaheads, preserving AND semantics.
func Regex(args []string) (string, error) {
	conditions, err := parseConditions(args)
	if err != nil {
		return "", err
	}

	for _, c := range conditions {
		if c.regexErr != nil {
			return "", &RegexError{
				Condition: c.source,
				Err:       c.regexErr,
			}
		}
	}

	var pattern strings.Builder
	pattern.WriteString(`\A`)

	for _, c := range conditions {
		pattern.WriteString(`(?=(?:`)
		pattern.WriteString(c.pattern)
		pattern.WriteString(`)\z)`)
	}

	pattern.WriteString(`(?s:.*)\z`)

	return pattern.String(), nil
}

func rangePattern(minimum, maximum *int) string {
	lower := 0
	if minimum != nil {
		lower = *minimum
	}

	if maximum == nil {
		return fmt.Sprintf(`(?s:.){%d,}`, lower)
	}

	return fmt.Sprintf(`(?s:.){%d,%d}`, lower, *maximum)
}
