// Package filter parses conditions and filters text streams.
package filter

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type condition struct {
	match    func(string) bool
	cost     int
	pattern  string
	regexErr error
	source   string
}

// New validates conditions and builds a filter. All conditions must match.
// Cheaper checks run first so rejected lines need less work.
func New(args []string) (*Filter, error) {
	conditions, err := parseConditions(args)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(conditions, func(i, j int) bool {
		return conditions[i].cost < conditions[j].cost
	})

	return &Filter{conditions: conditions}, nil
}

func parseConditions(args []string) ([]condition, error) {
	conditions := make([]condition, 0, len(args))

	for _, arg := range args {
		c, err := parseCondition(arg)
		if err != nil {
			return nil, &ConditionError{Condition: arg, Err: err}
		}

		c.source = arg
		conditions = append(conditions, c)
	}

	return conditions, nil
}

func parseCondition(raw string) (condition, error) {
	name, value, ok := strings.Cut(raw, "=")
	if !ok {
		return condition{}, errors.New("expected name=value")
	}

	switch name {
	case "range":
		return parseRange(value)
	case "format":
		return parseFormat(value)
	case "match":
		return parseTextMatch(value)
	case "initial", "charset":
		class, err := parseCharacterClass(value)
		if err != nil {
			return condition{}, err
		}

		if name == "initial" {
			return condition{
				cost:    2,
				pattern: class.pattern + `(?s:.*)`,
				match: func(s string) bool {
					for _, r := range s {
						return class.match(r)
					}

					return false
				},
			}, nil
		}

		return condition{
			cost:    3,
			pattern: class.pattern + `*`,
			match: func(s string) bool {
				for _, r := range s {
					if !class.match(r) {
						return false
					}
				}

				return true
			},
		}, nil
	default:
		return condition{}, fmt.Errorf("unknown condition name %q", name)
	}
}

func parseRange(raw string) (condition, error) {
	minRaw, maxRaw, ok := strings.Cut(raw, "..")
	if !ok {
		return condition{}, errors.New("range must have the form min..max")
	}

	if minRaw == "" && maxRaw == "" {
		return condition{}, errors.New("range requires a minimum or maximum")
	}

	minimum, err := parseBound("minimum", minRaw)
	if err != nil {
		return condition{}, err
	}

	maximum, err := parseBound("maximum", maxRaw)
	if err != nil {
		return condition{}, err
	}

	if minimum != nil && maximum != nil && *minimum > *maximum {
		return condition{}, errors.New("range minimum must be <= maximum")
	}

	var regexErr error
	if minimum != nil && *minimum > maxRegexRepetition || maximum != nil && *maximum > maxRegexRepetition {
		regexErr = fmt.Errorf("PCRE2 repetition bounds must be <= %d", maxRegexRepetition)
	}

	return condition{
		cost:     1,
		pattern:  rangePattern(minimum, maximum),
		regexErr: regexErr,
		match: func(s string) bool {
			length := utf8.RuneCountInString(s)

			return (minimum == nil || length >= *minimum) &&
				(maximum == nil || length <= *maximum)
		},
	}, nil
}

func parseBound(name, raw string) (*int, error) {
	if raw == "" {
		return nil, nil
	}

	n, err := strconv.Atoi(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid range %s %q: %w", name, raw, err)
	}

	if n < 0 {
		return nil, fmt.Errorf("range %s must be >= 0", name)
	}

	return &n, nil
}

type characterClass struct {
	match   func(rune) bool
	pattern string
}

func parseCharacterClass(raw string) (characterClass, error) {
	switch raw {
	case "alpha":
		return characterClass{match: unicode.IsLetter, pattern: `\p{L}`}, nil
	case "digit":
		return characterClass{match: unicode.IsDigit, pattern: `\p{Nd}`}, nil
	case "upper":
		return characterClass{match: unicode.IsUpper, pattern: `\p{Lu}`}, nil
	case "alnum":
		return characterClass{
			match: func(r rune) bool {
				return unicode.IsLetter(r) || unicode.IsDigit(r)
			},
			pattern: `[\p{L}\p{Nd}]`,
		}, nil
	default:
		return characterClass{}, fmt.Errorf("unknown character class %q", raw)
	}
}
