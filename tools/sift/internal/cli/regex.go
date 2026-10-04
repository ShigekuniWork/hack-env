package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"tools/sift/internal/filter"
)

func runRegex(args []string, stdout, stderr io.Writer) (err error) {
	options := newOptions(true, stderr)
	flags := options.flags
	outputPath := options.outputPath

	if parseErr := flags.Parse(args); parseErr != nil {
		if errors.Is(parseErr, flag.ErrHelp) {
			options.help()
			return nil
		}

		return &ArgumentError{Err: parseErr}
	}

	pattern, err := filter.Regex(flags.Args())
	if err != nil {
		return err
	}

	output := stdout
	if *outputPath != "-" {
		file, openErr := createOutput(*outputPath, nil)
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

	if _, err := fmt.Fprintln(output, pattern); err != nil {
		return &OutputError{Err: err}
	}

	return nil
}
