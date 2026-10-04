// Package cli handles sift command-line options and input/output files.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"tools/sift/internal/filter"
)

// Run handles command-line options and file ownership for the filter.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) (err error) {
	if len(args) > 0 && args[0] == "regex" {
		return runRegex(args[1:], stdout, stderr)
	}

	return runFilter(args, stdin, stdout, stderr)
}

func runFilter(args []string, stdin io.Reader, stdout, stderr io.Writer) (err error) {
	flags := flag.NewFlagSet("sift", flag.ContinueOnError)
	flags.SetOutput(stderr)

	inputPath := flags.String("input", "-", "input file (- for standard input)")
	outputPath := flags.String("output", "sift.txt", "output file (- for standard output)")

	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr,
			"Usage: sift [-input path] [-output path]"+
				" [range=min..max] [initial=alpha|digit|alnum|upper]"+
				" [charset=alpha|digit|alnum|upper] [format=url|email|domain] [match=pattern]",
		)
		flags.PrintDefaults()
		_, _ = fmt.Fprintln(stderr, "Generate a PCRE2 expression: sift regex [-output path] [conditions...]")
	}

	if parseErr := flags.Parse(args); parseErr != nil {
		if errors.Is(parseErr, flag.ErrHelp) {
			return nil
		}

		return &ArgumentError{Err: parseErr}
	}

	f, err := filter.New(flags.Args())
	if err != nil {
		return err
	}

	input := stdin
	var inputInfo os.FileInfo

	if *inputPath != "-" {
		file, openErr := os.Open(*inputPath)
		if openErr != nil {
			return &FileError{Op: "open input", Path: *inputPath, Err: openErr}
		}

		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				err = errors.Join(err, &FileError{Op: "close input", Path: *inputPath, Err: closeErr})
			}
		}()

		inputInfo, err = file.Stat()
		if err != nil {
			return &FileError{Op: "stat input", Path: *inputPath, Err: err}
		}

		input = file
	} else if file, ok := stdin.(*os.File); ok {
		inputInfo, err = file.Stat()
		if err != nil {
			return &FileError{Op: "stat input", Path: *inputPath, Err: err}
		}
	}

	output := stdout
	if *outputPath != "-" {
		file, openErr := createOutput(*outputPath, inputInfo)
		if openErr != nil {
			return openErr
		}

		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				err = errors.Join(err, &FileError{Op: "close output", Path: *outputPath, Err: closeErr})
			}
		}()

		output = file
	}

	return f.Run(input, output)
}
