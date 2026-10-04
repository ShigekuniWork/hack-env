package cli

import (
	"flag"
	"fmt"
	"io"
)

type commandOptions struct {
	flags      *flag.FlagSet
	inputPath  *string
	outputPath *string
	stderr     io.Writer
	regex      bool
}

func newOptions(regex bool, stderr io.Writer) *commandOptions {
	name, outputDefault := "sift", "sift.txt"
	if regex {
		name, outputDefault = "sift regex", "-"
	}

	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Usage = func() {}

	options := &commandOptions{
		flags:      flags,
		outputPath: flags.String("output", outputDefault, "output file (- for standard output)"),
		stderr:     stderr,
		regex:      regex,
	}
	if !regex {
		options.inputPath = flags.String("input", "-", "input file (- for standard input)")
	}

	return options
}

func (o *commandOptions) help() {
	usage := "Usage: sift [-input path] [-output path] [conditions...]"
	if o.regex {
		usage = "Usage: sift regex [-output path] [conditions...]"
	}

	_, _ = fmt.Fprintln(o.stderr, usage)
	_, _ = fmt.Fprintln(o.stderr, "Options:")
	o.flags.SetOutput(o.stderr)
	o.flags.PrintDefaults()
	_, _ = fmt.Fprintln(o.stderr, "  -h, --help\n    show help")
	_, _ = fmt.Fprintln(o.stderr, `
Conditions (AND):
  range=min..max                 Unicode character count; either bound may be omitted
  initial=alpha|digit|alnum|upper First character class
  charset=alpha|digit|alnum|upper Character class for the entire line
  format=url|email|domain        Built-in format check
  match=pattern                 abc: exact, abc*: prefix, *abc: suffix, *abc*: contains

Quote patterns containing * to prevent shell expansion.`)
	if !o.regex {
		_, _ = fmt.Fprintln(o.stderr, "\nGenerate a PCRE2 expression: sift regex [-output path] [conditions...]")
	}
}

// PrintError displays a failure followed by help for the selected command.
func PrintError(args []string, err error, stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, "error:", err)
	_, _ = fmt.Fprintln(stderr)
	regex := len(args) > 0 && args[0] == "regex"
	newOptions(regex, stderr).help()
}
