package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"tools/sift/internal/filter"
)

func TestRegexOutput(t *testing.T) {
	t.Chdir(t.TempDir())
	var stdout bytes.Buffer
	want := `\A(?=(?:Ab(?s:.*))\z)(?s:.*)\z` + "\n"

	if err := Run([]string{"regex", "match=Ab*"}, unreadableInput{}, &stdout, io.Discard); err != nil {
		t.Fatal(err)
	}

	if stdout.String() != want {
		t.Fatalf("output = %q, want %q", stdout.String(), want)
	}

	if _, err := os.Stat("sift.txt"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("regex command created default filter output")
	}

	stdout.Reset()
	if err := Run([]string{"regex", "-output", "pattern.txt", "match=Ab*"}, unreadableInput{}, &stdout, io.Discard); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile("pattern.txt")
	if err != nil || string(got) != want || stdout.Len() != 0 {
		t.Fatalf("file output = %q, stdout = %q, error = %v", got, stdout.String(), err)
	}
}

type unreadableInput struct{}

func (unreadableInput) Read([]byte) (int, error) {
	panic("regex generation must not read input")
}

func TestRegexErrorsAndHelp(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("pattern.txt", []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, raw := range []string{"format=unknown", "match=a*b", "range=65536.."} {
		var stdout bytes.Buffer
		if err := Run([]string{"regex", "-output", "pattern.txt", raw}, unreadableInput{}, &stdout, io.Discard); err == nil {
			t.Fatalf("expected error for %q", raw)
		}

		got, err := os.ReadFile("pattern.txt")
		if err != nil || string(got) != "keep" || stdout.Len() != 0 {
			t.Fatalf("invalid conditions changed output: %q, %v", got, err)
		}
	}

	var help bytes.Buffer
	if err := Run([]string{"regex", "-h"}, unreadableInput{}, io.Discard, &help); err != nil || !strings.Contains(help.String(), "sift regex") {
		t.Fatalf("help = %q, error = %v", help.String(), err)
	}

	err := Run([]string{"regex", "-input", "missing"}, unreadableInput{}, io.Discard, io.Discard)
	var argumentErr *ArgumentError
	if !errors.As(err, &argumentErr) {
		t.Fatalf("expected ArgumentError: %v", err)
	}

	err = Run([]string{"regex", "format=unknown"}, unreadableInput{}, io.Discard, io.Discard)
	var conditionErr *filter.ConditionError
	if !errors.As(err, &conditionErr) {
		t.Fatalf("expected ConditionError: %v", err)
	}
}

type failingOutput struct{ err error }

func (w failingOutput) Write([]byte) (int, error) {
	return 0, w.err
}

func TestRegexWriteError(t *testing.T) {
	failure := errors.New("write failed")
	err := Run([]string{"regex"}, unreadableInput{}, failingOutput{failure}, io.Discard)

	var outputErr *OutputError
	if !errors.Is(err, failure) || !errors.As(err, &outputErr) {
		t.Fatalf("unexpected write error: %v", err)
	}
}
