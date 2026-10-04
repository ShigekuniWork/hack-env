package filter

import (
	"fmt"
	"regexp"
)

const (
	domainLabelPattern = `[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?`
	domainPattern      = `(?:` + domainLabelPattern + `\.)+[A-Za-z](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?`
	emailPattern       = `[A-Za-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[A-Za-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+)*@` + domainPattern
	urlPattern         = `(?i:https?)://` + domainLabelPattern + `(?:\.` + domainLabelPattern + `)*(?::[0-9]{1,5})?(?:[/?#][^ \t\n\f\r]*)?`
)

var (
	domainCondition = formatCondition(domainPattern)
	emailCondition  = formatCondition(emailPattern)
	urlCondition    = formatCondition(urlPattern)
)

func parseFormat(value string) (condition, error) {
	switch value {
	case "url":
		return urlCondition, nil
	case "email":
		return emailCondition, nil
	case "domain":
		return domainCondition, nil
	default:
		return condition{}, fmt.Errorf("unknown format %q", value)
	}
}

func formatCondition(pattern string) condition {
	compiled := regexp.MustCompile(`\A` + pattern + `\z`)

	return condition{
		cost:    4,
		match:   compiled.MatchString,
		pattern: pattern,
	}
}
