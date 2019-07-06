package xtrace

import (
	"strings"
)

// TraceFormatter allows for the formatting of any message in the given trace
type TraceFormatter interface {
	// FormatTrace takes a message and will return a formatted message. Any previous message many be updated based on
	// the contents of the passed message
	FormatTrace(previousMessages []string, message string) string
}

// NilFormatter applies no formatting and returns the given message as is
type NilFormatter struct{}

// FormatTrace applies no formatting and returns the given message as is
func (formatter NilFormatter) FormatTrace(previousMessages []string, message string) string {
	return message
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
