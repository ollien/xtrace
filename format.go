package xtrace

import (
	"regexp"
	"strings"
)

// TraceFormatter allows for the formatting of any message in the given trace
type TraceFormatter interface {
	// FormatTrace takes a message and will return a formatted message. Any previous message many be updated based on
	// the contents of the passed message
	FormatTrace(previousMessages []string, message string) string
}

// NilFormatter applies no formatting and returns the given message as xerrors sends them.
// Note that the messages that xerrors sends aren't always the most intuitive (e.g. there are no newlines after error
// messages), and the usage of this formatter is not strictly recommended. It is mainly provided for those that want
// a simple shim.
type NilFormatter struct{}

// FormatTrace applies no formatting and returns the given message as is
func (formatter NilFormatter) FormatTrace(previousMessages []string, message string) string {
	return message
}

// NewLineFormatter ensures that all messages except the last end in a newline after all error content.
type NewLineFormatter struct {
	// Naive, if set, will instruct the formatter to perform the naive version of this algorithm, which simply
	// adds/removes a newline from the end of each message. xerrors has a habit of sending indentation in the previous
	// line (i.e. "<error>\n    "), so the naive algorithm produces weird output. Nevertheless, this may be desirable
	// depending on the implementation of xerrors.Formatter, so it is left as an option.
	Naive bool
	// holds the last message with no newline stripped
	lastRawMessage string
}

// FormatTrace formats the message as dictated by the contract for NewLineFormatter
func (formatter *NewLineFormatter) FormatTrace(previousMessages []string, message string) (formatted string) {
	lastMessage := formatter.lastRawMessage
	formatter.lastRawMessage = message
	formatted = formatter.stripNewlines(message)
	if len(previousMessages) == 0 {
		return
	}

	// Add a newline back to the last message
	terminatedLastMessage := formatter.newLineTerminateMessage(lastMessage)
	previousMessages[len(previousMessages)-1] = terminatedLastMessage

	return
}

func (formatter *NewLineFormatter) stripNewlines(message string) string {
	if formatter.Naive {
		return strings.TrimRight(message, "\n")
	}

	// Capture group will only match the trailing whitespace portion of the string
	pattern := regexp.MustCompile(`.*\S(\s*)`)
	matchBoundaries := pattern.FindStringSubmatchIndex(message)
	// If we don't match, we don't need to strip anything
	if matchBoundaries == nil {
		return message
	}

	// Remove all newlines from the whitespace portion
	// The capture group's boundaries will be stored at the 2nd and 3rd positions (zero-indexed)
	errorPortion, whitespacePortion := message[:matchBoundaries[2]], message[matchBoundaries[2]:matchBoundaries[3]]
	strippedWhitespacePortion := strings.Replace(whitespacePortion, "\n", "", -1)

	return errorPortion + strippedWhitespacePortion
}

func (formatter *NewLineFormatter) newLineTerminateMessage(message string) string {
	pattern := regexp.MustCompile(`\s*\n\s*$`)
	// Make sure the previous message ends with a newline, or there is newline within a trailing whitespace region.
	if (formatter.Naive && message[len(message)-1] == '\n') ||
		(!formatter.Naive && pattern.MatchString(message)) {
		return message
	}

	return message + "\n"
}

// NestedMessageFormatter will leave the leading line with no indentation, but indents all lines following, stripping
// whitespace from the both the left and right of each line and replacing it with a newline, unless it is the last
// message. In this case, no newline is inserted, but whitespace is still stripped.
type NestedMessageFormatter struct {
	Indentation string
}

// FormatTrace formats the message as dictated by the contract for NestedMessageFormatter
func (formatter NestedMessageFormatter) FormatTrace(previousMessages []string, message string) string {
	formattedMessage := strings.TrimSpace(message)
	// All messages except the first must begin with the given indentation, so if we have the first, we're done.
	if len(previousMessages) == 0 {
		return formattedMessage
	}

	formattedMessage = formatter.Indentation + formattedMessage
	lastMessage := previousMessages[len(previousMessages)-1]
	// Make sure the previous message ends with a newline
	if lastMessage[len(lastMessage)-1] != '\n' {
		lastMessage += "\n"
		previousMessages[len(previousMessages)-1] = lastMessage
	}

	return formattedMessage
}
