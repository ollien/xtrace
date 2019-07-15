package xtrace

// Naive will set the naive flag when passed to NewNewLineFormatter. This flag, if set, will instruct the formatter
// to perform the naive version of this algorithm, which simply adds/removes a newline from the end of each message.
// xerrors has a habit of sending indentation in the previous line (i.e. "<error>\n    "), so the naive algorithm
// produces weird output. Nevertheless, this may be desirable depending on the implementation of xerrors.Formatter, so
// it is left as an option.
func Naive(naive bool) func(*NewLineFormatter) error {
	return func(formatter *NewLineFormatter) error {
		formatter.naive = naive

		return nil
	}
}

// NestingIndentation sets the indentation for the NestedMessageFormatter that is produced when this is passed to
// NewNestedMessageFormatter.
func NestingIndentation(indentation string) func(*NestedMessageFormatter) error {
	return func(formatter *NestedMessageFormatter) error {
		formatter.indentation = indentation

		return nil
	}
}
