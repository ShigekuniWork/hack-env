package filter

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"
)

func TestConditions(t *testing.T) {
	tests := []struct {
		name string
		args []string
		line string
		want bool
	}{
		{"no conditions", nil, "", true},
		{"unicode range", []string{"range=2..3"}, "日本語", true},
		{"range too short", []string{"range=2..3"}, "日", false},
		{"range too long", []string{"range=2..3"}, "abcd", false},
		{"minimum only", []string{"range=2.."}, "abc", true},
		{"maximum only", []string{"range=..2"}, "", true},
		{"zero", []string{"range=0..0"}, "", true},
		{"unicode initial", []string{"initial=alpha"}, "日1!", true},
		{"digit initial", []string{"initial=digit"}, "１abc", true},
		{"empty initial", []string{"initial=alnum"}, "", false},
		{"uppercase initial", []string{"initial=upper"}, "Apple", true},
		{"unicode uppercase initial", []string{"initial=upper"}, "Éclair", true},
		{"lowercase initial", []string{"initial=upper"}, "apple", false},
		{"uncased initial", []string{"initial=upper"}, "日本語", false},
		{"digit is not uppercase", []string{"initial=upper"}, "1A", false},
		{"empty uppercase initial", []string{"initial=upper"}, "", false},
		{"uppercase charset", []string{"charset=upper"}, "AÉ", true},
		{"mixed case charset", []string{"charset=upper"}, "Aa", false},
		{"unicode charset", []string{"charset=alpha"}, "日本語", true},
		{"digits", []string{"charset=digit"}, "１２3", true},
		{"alphanumeric", []string{"charset=alnum"}, "日１a", true},
		{"punctuation", []string{"charset=alnum"}, "a!", false},
		{"empty charset", []string{"charset=alpha"}, "", true},
		{"all conditions", []string{"charset=alnum", "range=2..3", "initial=alpha"}, "日1", true},
		{"one fails", []string{"charset=alnum", "range=2..3", "initial=alpha"}, "12", false},
		{"repeated conditions", []string{"range=2..", "range=..3"}, "abcd", false},
		{"format and text", []string{"format=email", "match=*example.com", "initial=upper"}, "Alice@example.com", true},
		{"format and text fail", []string{"format=email", "match=*example.com"}, "alice@other.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := New(tt.args)
			if err != nil {
				t.Fatal(err)
			}

			if got := f.Match(tt.line); got != tt.want {
				t.Fatalf("Match(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestInvalidConditions(t *testing.T) {
	invalidConditions := []string{
		"range",
		"unknown=x",
		"range=1",
		"range=..",
		"range=-1..2",
		"range=1..-2",
		"range=3..2",
		"range=x..2",
		"range=1..x",
		"range=1..2..3",
		"initial=unknown",
		"charset=",
		"format=",
		"format=unknown",
		"match=a*b",
		"match=***",
		"match=**abc",
		"match=abc**",
	}

	for _, raw := range invalidConditions {
		t.Run(raw, func(t *testing.T) {
			_, err := New([]string{raw})

			var conditionErr *ConditionError

			if !errors.As(err, &conditionErr) || conditionErr.Condition != raw {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	_, err := New([]string{"range=x..2"})

	var numberErr *strconv.NumError

	if !errors.As(err, &numberErr) {
		t.Fatalf("missing underlying number error: %v", err)
	}
}

func TestRun(t *testing.T) {
	f, err := New([]string{"range=2..3", "charset=alpha"})
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer

	if err := f.Run(strings.NewReader("a\n日本語\r\n12\nabc"), &out); err != nil {
		t.Fatal(err)
	}

	if got := out.String(); got != "日本語\nabc\n" {
		t.Fatalf("output = %q", got)
	}
}

type brokenReader struct{ err error }

func (r brokenReader) Read([]byte) (int, error) {
	return 0, r.err
}

type brokenWriter struct{ err error }

func (w brokenWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestStreamErrors(t *testing.T) {
	failure := errors.New("stream failed")
	f := &Filter{}
	tests := []struct {
		name   string
		input  io.Reader
		output io.Writer
		op     string
	}{
		{"read", brokenReader{failure}, io.Discard, "read input"},
		{"flush", strings.NewReader("short\n"), brokenWriter{failure}, "flush output"},
		{"write", strings.NewReader(strings.Repeat("a", 8192)), brokenWriter{failure}, "write output"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := f.Run(tt.input, tt.output)

			var streamErr *StreamError

			if !errors.Is(err, failure) || !errors.As(err, &streamErr) || streamErr.Op != tt.op {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
