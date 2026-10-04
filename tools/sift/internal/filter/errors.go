package filter

import "fmt"

// ConditionError identifies the condition that could not be parsed or validated.
type ConditionError struct {
	Condition string
	Err       error
}

func (e *ConditionError) Error() string {
	return fmt.Sprintf("invalid condition %q: %v", e.Condition, e.Err)
}

func (e *ConditionError) Unwrap() error {
	return e.Err
}

// RegexError identifies a valid condition that cannot be represented in PCRE2.
type RegexError struct {
	Condition string
	Err       error
}

func (e *RegexError) Error() string {
	return fmt.Sprintf("cannot generate regex for %q: %v", e.Condition, e.Err)
}

func (e *RegexError) Unwrap() error {
	return e.Err
}

// StreamError identifies the failed input or output operation.
type StreamError struct {
	Op  string
	Err error
}

func (e *StreamError) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *StreamError) Unwrap() error {
	return e.Err
}
