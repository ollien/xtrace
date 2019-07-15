package xtrace

/**
  Copyright 2019 Nicholas Krichevsky

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"regexp"
	"strings"

	"golang.org/x/xerrors"
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

// NewNilFormatter makes a new NilFormatter
func NewNilFormatter() *NilFormatter {
	return &NilFormatter{}
}

// FormatTrace applies no formatting and returns the given message as is
func (formatter NilFormatter) FormatTrace(previousMessages []string, message string) string {
	return message
}

// NestedMessageFormatter will leave the leading line with no indentation, but indents all lines following, stripping
// whitespace from the both the left and right of each line and replacing it with a newline, unless it is the last
// message. In this case, no newline is inserted, but whitespace is still stripped.
type NestedMessageFormatter struct {
	indentation string
}

// NewNestedMessageFormatter makes a new NestedMessageFormatter
func NewNestedMessageFormatter(options ...func(*NestedMessageFormatter) error) (*NestedMessageFormatter, error) {
	formatter := &NestedMessageFormatter{}
	for _, optionFunc := range options {
		err := optionFunc(formatter)
		if err != nil {
			return nil, xerrors.Errorf("Could not construct NestedMessageFormatter: %w", err)
		}
	}

	return formatter, nil
}

// FormatTrace formats the message as dictated by the contract for NestedMessageFormatter
func (formatter NestedMessageFormatter) FormatTrace(previousMessages []string, message string) string {
	formattedMessage := strings.TrimSpace(message)
	// All messages except the first must begin with the given indentation, so if we have the first, we're done.
	if len(previousMessages) == 0 {
		return formattedMessage
	}

	formattedMessage = formatter.indentation + formattedMessage
	lastMessage := previousMessages[len(previousMessages)-1]
	// Make sure the previous message ends with a newline
	if lastMessage[len(lastMessage)-1] != '\n' {
		lastMessage += "\n"
		previousMessages[len(previousMessages)-1] = lastMessage
	}

	return formattedMessage
}

// NewLineFormatter ensures that all messages except the last end in a newline after all error content.
type NewLineFormatter struct {
	// naive will enable the naive algorithm. See the Naive method for more info
	naive bool
	// holds the last message with no newline stripped
	lastRawMessage string
}

// NewNewLineFormatter will make a new NewLineFormatter
func NewNewLineFormatter(options ...func(*NewLineFormatter) error) (*NewLineFormatter, error) {
	formatter := &NewLineFormatter{}
	for _, optionFunc := range options {
		err := optionFunc(formatter)
		if err != nil {
			return nil, xerrors.Errorf("Could not construct NewLineFormatter: %w", err)
		}
	}

	return formatter, nil
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

// stripNewLines will strip new lines from the message using the given strategy
func (formatter *NewLineFormatter) stripNewlines(message string) string {
	if formatter.naive {
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

// newLineTerminateMessages will termiante the message with a newline, based on the given strategy
func (formatter *NewLineFormatter) newLineTerminateMessage(message string) string {
	pattern := regexp.MustCompile(`\s*\n\s*$`)
	// Make sure the previous message ends with a newline, or there is newline within a trailing whitespace region.
	if (formatter.naive && message[len(message)-1] == '\n') ||
		(!formatter.naive && pattern.MatchString(message)) {
		return message
	}

	return message + "\n"
}
