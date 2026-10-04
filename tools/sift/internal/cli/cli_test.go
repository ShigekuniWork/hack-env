package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tools/sift/internal/filter"
)

func TestDefaultOutput(t *testing.T) {
	t.Chdir(t.TempDir())

	var stdout bytes.Buffer

	if err := Run([]string{"charset=alpha"}, strings.NewReader("abc\n123\n日本語\n"), &stdout, io.Discard); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile("sift.txt")
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "abc\n日本語\n" || stdout.Len() != 0 {
		t.Fatalf("file = %q, stdout = %q", got, stdout.String())
	}
}

func TestFileInputAndOutput(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	input := filepath.Join(dir, "input.txt")
	output := filepath.Join(dir, "output.txt")
	if err := os.WriteFile(input, []byte("a\nabc\n123\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(output, []byte("old output"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := Run([]string{"--input", input, "--output", output, "range=2..", "initial=alpha"}, strings.NewReader(""), io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile("output.txt")
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "abc\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestStandardStreams(t *testing.T) {
	var out bytes.Buffer

	if err := Run([]string{"-input", "-", "-output", "-", "initial=digit"}, strings.NewReader("abc\n123\n"), &out, io.Discard); err != nil {
		t.Fatal(err)
	}

	if out.String() != "123\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestCombinedMatchingConditions(t *testing.T) {
	var out bytes.Buffer

	args := []string{"-output", "-", "format=email", "initial=upper", "match=*example.com"}
	input := "Alice@example.com\nalice@example.com\nBob@other.com\nNotAnEmail@example.com/path\n"
	if err := Run(args, strings.NewReader(input), &out, io.Discard); err != nil {
		t.Fatal(err)
	}

	if got := out.String(); got != "Alice@example.com\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestSameFileIsPreserved(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	input := filepath.Join(dir, "input.txt")
	alias := filepath.Join(dir, "alias.txt")
	content := "keep this\n"
	if err := os.WriteFile(input, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Link(input, alias); err != nil {
		t.Fatal(err)
	}

	for _, output := range []string{input, alias} {
		err := Run([]string{"-input", input, "-output", output}, strings.NewReader(""), io.Discard, io.Discard)

		var fileErr *FileError

		if !errors.As(err, &fileErr) {
			t.Fatalf("expected FileError, got %v", err)
		}

		got, readErr := os.ReadFile("input.txt")
		if readErr != nil || string(got) != content {
			t.Fatalf("input changed: %q, %v", got, readErr)
		}
	}

	file, err := os.Open("input.txt")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Error(err)
		}
	})

	if err := Run([]string{"-output", alias}, file, io.Discard, io.Discard); err == nil {
		t.Fatal("expected redirected stdin conflict")
	}
}

func TestInvalidConditionDoesNotOverwrite(t *testing.T) {
	t.Chdir(t.TempDir())

	output := "out.txt"
	if err := os.WriteFile(output, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, raw := range []string{"range=bad", "format=unknown", "match=a*b"} {
		err := Run([]string{"-output", output, raw}, strings.NewReader(""), io.Discard, io.Discard)

		var conditionErr *filter.ConditionError

		if !errors.As(err, &conditionErr) {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := os.ReadFile("out.txt")
		if err != nil || string(got) != "keep" {
			t.Fatalf("output changed: %q, %v", got, err)
		}
	}
}

func TestMissingInput(t *testing.T) {
	err := Run([]string{"-input", filepath.Join(t.TempDir(), "missing"), "-output", "-"}, strings.NewReader(""), io.Discard, io.Discard)

	var fileErr *FileError

	if !errors.Is(err, os.ErrNotExist) || !errors.As(err, &fileErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOptions(t *testing.T) {
	for _, args := range [][]string{{"-unknown"}, {"-input"}} {
		err := Run(args, strings.NewReader(""), io.Discard, io.Discard)

		var argumentErr *ArgumentError

		if !errors.As(err, &argumentErr) {
			t.Fatalf("expected ArgumentError, got %v", err)
		}
	}

	t.Chdir(t.TempDir())

	var help bytes.Buffer

	if err := Run([]string{"-h"}, strings.NewReader(""), io.Discard, &help); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(help.String(), "sift.txt") {
		t.Fatal("missing help")
	}

	if _, err := os.Stat("sift.txt"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("help created output")
	}
}
