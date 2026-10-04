package filter

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func TestRegex(t *testing.T) {
	pattern, err := Regex([]string{"initial=upper", "match=Ab*"})
	if err != nil {
		t.Fatal(err)
	}

	want := `\A(?=(?:\p{Lu}(?s:.*))\z)(?=(?:Ab(?s:.*))\z)(?s:.*)\z`
	if pattern != want {
		t.Fatalf("Regex = %q, want %q", pattern, want)
	}

	pattern, err = Regex(nil)
	if err != nil || pattern != `\A(?s:.*)\z` {
		t.Fatalf("empty conditions: %q, %v", pattern, err)
	}
}

func TestRegexFragmentsAgreeWithFilter(t *testing.T) {
	conditions := []string{
		"range=0..0", "range=2..3", "range=2..", "range=..2",
		"initial=upper", "initial=alpha", "initial=digit", "initial=alnum",
		"charset=upper", "charset=alpha", "charset=digit", "charset=alnum",
		"format=url", "format=email", "format=domain",
		"match=abc", "match=abc*", "match=*abc", "match=*abc*",
		"match=", "match=*", "match=**", "match=a.b[0]", "match=*日本*",
	}
	lines := []string{
		"", "abc", "ABC", "xabc", "abcx", "xabcx", "a.b[0]", "日本語", "Éclair", "１２3",
		"a\nb", "https://example.com", "Alice@example.com", "example.com", "example.com\n",
	}

	for _, raw := range conditions {
		t.Run(raw, func(t *testing.T) {
			c, err := parseCondition(raw)
			if err != nil {
				t.Fatal(err)
			}

			pattern, err := regexp.Compile(`\A(?:` + c.pattern + `)\z`)
			if err != nil {
				t.Fatal(err)
			}

			for _, line := range lines {
				if got, want := pattern.MatchString(line), c.match(line); got != want {
					t.Errorf("regex and check disagree for %q: regex=%v, check=%v", line, got, want)
				}
			}
		})
	}
}

func TestRegexRangeLimit(t *testing.T) {
	_, err := Regex([]string{"range=65536.."})

	var regexErr *RegexError
	if !errors.As(err, &regexErr) || regexErr.Condition != "range=65536.." || errors.Unwrap(regexErr) == nil {
		t.Fatalf("unexpected generation error: %v", err)
	}

	f, err := New([]string{"range=65536.."})
	if err != nil || !f.Match(strings.Repeat("a", 65536)) {
		t.Fatalf("normal filtering should still support large ranges: %v", err)
	}

	if _, err := Regex([]string{"range=..65535"}); err != nil {
		t.Fatal(err)
	}
}

func TestRegexPCRE2(t *testing.T) {
	_, err := exec.LookPath("rg")
	if err != nil {
		t.Skip("ripgrep is unavailable")
	}

	if err := exec.CommandContext(t.Context(), "rg", "--pcre2-version").Run(); err != nil {
		t.Skip("ripgrep has no PCRE2 support")
	}

	t.Chdir(t.TempDir())

	conditions := [][]string{
		nil,
		{"initial=upper", "charset=alpha", "range=2..8", "match=*clair"},
		{"format=email", "match=*example.com", "initial=upper"},
		{"format=url", "match=https*"},
		{"format=domain", "match=*example*"},
		{"charset=alnum", "initial=digit", "range=1..3"},
		{"match=a.b[0]"},
		{"match=", "range=0..0"},
	}
	lines := []string{"", "Éclair", "éclair", "Alice@example.com", "alice@example.com", "Bob@other.com", "https://example.com", "example.com", "１２3", "a.b[0]"}

	for _, args := range conditions {
		pattern, err := Regex(args)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile("pattern.txt", []byte(pattern+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}

		f, err := New(args)
		if err != nil {
			t.Fatal(err)
		}

		for _, line := range lines {
			cmd := exec.CommandContext(t.Context(), "rg", "--pcre2", "--quiet", "--file", "pattern.txt")
			cmd.Stdin = strings.NewReader(line + "\n")
			output, err := cmd.CombinedOutput()

			var exitErr *exec.ExitError
			if err != nil && (!errors.As(err, &exitErr) || exitErr.ExitCode() != 1) {
				t.Fatalf("PCRE2 rejected %q: %v: %s", pattern, err, output)
			}

			if got, want := err == nil, f.Match(line); got != want {
				t.Errorf("%v on %q: PCRE2=%v, filter=%v", args, line, got, want)
			}
		}
	}
}
