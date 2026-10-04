package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestErrorHelp(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		usage string
	}{
		{"unknown flag", []string{"-unknown"}, "Usage: sift ["},
		{"missing option value", []string{"-input"}, "Usage: sift ["},
		{"invalid condition", []string{"format=unknown"}, "Usage: sift ["},
		{"missing input file", []string{"-input", "missing.txt", "-output", "-"}, "Usage: sift ["},
		{"regex flag", []string{"regex", "-input", "missing.txt"}, "Usage: sift regex ["},
		{"regex condition", []string{"regex", "match=a*b"}, "Usage: sift regex ["},
		{"regex generation", []string{"regex", "range=65536.."}, "Usage: sift regex ["},
	}

	t.Chdir(t.TempDir())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			err := Run(tt.args, strings.NewReader(""), io.Discard, &stderr)
			if err == nil {
				t.Fatal("expected command failure")
			}

			PrintError(tt.args, err, &stderr)
			got := stderr.String()
			if strings.Count(got, "error:") != 1 || strings.Count(got, "Usage:") != 1 {
				t.Fatalf("diagnostics should appear once: %q", got)
			}

			if !strings.Contains(got, tt.usage) || !strings.Contains(got, "format=url|email|domain") || !strings.Contains(got, "*abc*: contains") {
				t.Fatalf("missing command help: %q", got)
			}

			if strings.Index(got, "error:") > strings.Index(got, "Usage:") {
				t.Fatalf("help precedes the error: %q", got)
			}
		})
	}
}

func TestNoArgumentsWithCharacterDevice(t *testing.T) {
	t.Chdir(t.TempDir())
	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := stdin.Close(); err != nil {
			t.Error(err)
		}
	})

	var stderr bytes.Buffer
	if err := Run(nil, stdin, io.Discard, &stderr); err != nil {
		t.Fatal(err)
	}

	if strings.Count(stderr.String(), "Usage:") != 1 {
		t.Fatalf("missing help: %q", stderr.String())
	}

	if _, err := os.Stat("sift.txt"); !os.IsNotExist(err) {
		t.Fatal("help created an output file")
	}
}

func TestNoArgumentsWithPipedInput(t *testing.T) {
	t.Chdir(t.TempDir())
	stdin, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := stdin.Close(); err != nil {
			t.Error(err)
		}
	})

	content := "hello\n日本語\n"
	if _, err := io.WriteString(writer, content); err != nil {
		_ = writer.Close()
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	if err := Run(nil, stdin, io.Discard, &stderr); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile("sift.txt")
	if err != nil || string(got) != content || stderr.Len() != 0 {
		t.Fatalf("file = %q, stderr = %q, error = %v", got, stderr.String(), err)
	}
}
