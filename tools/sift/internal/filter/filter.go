package filter

import (
	"bufio"
	"fmt"
	"io"
)

// Filter selects lines using validated conditions. Its zero value matches all lines.
type Filter struct {
	conditions []condition
}

// Match reports whether a line satisfies every condition.
func (f *Filter) Match(line string) bool {
	for _, c := range f.conditions {
		if !c.match(line) {
			return false
		}
	}

	return true
}

// Run filters input line by line and flushes matching lines to output.
// The caller owns the streams and is responsible for closing them.
func (f *Filter) Run(input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	writer := bufio.NewWriter(output)

	for scanner.Scan() {
		line := scanner.Text()
		if f.Match(line) {
			if _, err := fmt.Fprintln(writer, line); err != nil {
				return &StreamError{Op: "write output", Err: err}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return &StreamError{Op: "read input", Err: err}
	}

	if err := writer.Flush(); err != nil {
		return &StreamError{Op: "flush output", Err: err}
	}

	return nil
}
