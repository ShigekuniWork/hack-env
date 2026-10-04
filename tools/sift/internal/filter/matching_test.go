package filter

import (
	"strings"
	"testing"
)

func TestFormats(t *testing.T) {
	tests := []struct {
		format  string
		valid   []string
		invalid []string
	}{
		{
			format: "url",
			valid: []string{
				"https://example.com", "HTTP://Example.COM/path?q=1#section",
				"http://localhost:8080/", "https://127.0.0.1:443/a",
			},
			invalid: []string{
				"", "example.com", "ftp://example.com", "https://", "https://-example.com",
				"https://example.com:abc", "https://example.com/a b", "prefix https://example.com",
				"https://example.com\n", "https://user@example.com", "https://日本.jp",
			},
		},
		{
			format: "email",
			valid: []string{
				"alice@example.com", "first.last+tag@sub.example.co.jp", "a@EXAMPLE.COM",
			},
			invalid: []string{
				"", "alice", "@example.com", "alice@localhost", ".alice@example.com",
				"alice.@example.com", "alice..bob@example.com", "alice@example..com",
				"alice@-example.com", "alice@example.com\n", " alice@example.com", "あ@example.com",
			},
		},
		{
			format: "domain",
			valid: []string{
				"example.com", "sub.example.co.jp", "EXAMPLE.COM", "xn--wgv71a.jp",
				strings.Repeat("a", 63) + ".com",
			},
			invalid: []string{
				"", "localhost", "https://example.com", "example.com/path", "a@example.com",
				"-example.com", "example-.com", "example..com", "example.com.", "example_com.jp",
				"127.0.0.1", "日本.jp", "example.com\n", strings.Repeat("a", 64) + ".com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			f, err := New([]string{"format=" + tt.format})
			if err != nil {
				t.Fatal(err)
			}

			for _, line := range tt.valid {
				if !f.Match(line) {
					t.Errorf("Match(%q) = false, want true", line)
				}
			}

			for _, line := range tt.invalid {
				if f.Match(line) {
					t.Errorf("Match(%q) = true, want false", line)
				}
			}
		})
	}
}

func TestTextMatch(t *testing.T) {
	tests := []struct {
		pattern string
		line    string
		want    bool
	}{
		{"abc", "abc", true},
		{"abc", "abcd", false},
		{"abc", "ABC", false},
		{"abc*", "abcdef", true},
		{"abc*", "xabc", false},
		{"*abc", "xyzabc", true},
		{"*abc", "abcx", false},
		{"*abc*", "xyzabcdef", true},
		{"*abc*", "ab", false},
		{"日本*", "日本語", true},
		{"a.b[0]", "a.b[0]", true},
		{"a.b[0]", "axb0", false},
		{"a=b", "a=b", true},
		{"", "", true},
		{"", "abc", false},
		{"*", "", true},
		{"*", "abc", true},
		{"**", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"/"+tt.line, func(t *testing.T) {
			f, err := New([]string{"match=" + tt.pattern})
			if err != nil {
				t.Fatal(err)
			}

			if got := f.Match(tt.line); got != tt.want {
				t.Fatalf("Match(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}
