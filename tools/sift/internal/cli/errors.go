package cli

import "fmt"

// OutputError preserves a failure to write command output.
type OutputError struct {
	Err error
}

func (e *OutputError) Error() string {
	return fmt.Sprintf("write output: %v", e.Err)
}

func (e *OutputError) Unwrap() error {
	return e.Err
}

// ArgumentError reports invalid command-line options.
type ArgumentError struct {
	Err error
}

func (e *ArgumentError) Error() string {
	return fmt.Sprintf("invalid arguments: %v", e.Err)
}

func (e *ArgumentError) Unwrap() error {
	return e.Err
}

// FileError adds the operation and path while preserving the underlying error.
type FileError struct {
	Op   string
	Path string
	Err  error
}

func (e *FileError) Error() string {
	return fmt.Sprintf("%s %q: %v", e.Op, e.Path, e.Err)
}

func (e *FileError) Unwrap() error {
	return e.Err
}
