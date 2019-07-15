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
	"bytes"
	"fmt"
	"io"

	"golang.org/x/xerrors"
)

const emptyError = "<empty>"

// Tracer gets the trace of an error
type Tracer struct {
	detailedOutput bool
	// Populated with the full chain of errors, with the originating error at len(errorChain) - 1
	errorChain []error
	// Holds the contents of the current error being read
	buffer *bytes.Buffer
	// Formats the traces returned by the Read functions
	formatter TraceFormatter
	// Sets the order of the method
	ordering TraceOrderingMethod
	// baseError is the original error passed, primarily used for cloning purposes
	baseErr error
	// holds all of the option functions passed to the tracer, primarily used for cloning purposes
	optionFuncs []func(*Tracer) error
}

// NewTracer returns a new tracer for the given error
func NewTracer(baseErr error, options ...func(*Tracer) error) (*Tracer, error) {
	formatter, err := NewNewLineFormatter(Naive(false))
	if err != nil {
		return nil, xerrors.Errorf("Could not construct formatter for Tracer: %w")
	}

	tracer := &Tracer{
		errorChain:     buildErrorChain(baseErr),
		detailedOutput: true,
		buffer:         bytes.NewBuffer([]byte{}),
		formatter:      formatter,
		ordering:       OldestFirstOrdering,
		baseErr:        baseErr,
		optionFuncs:    options,
	}

	for _, optionFunc := range options {
		err := optionFunc(tracer)
		if err != nil {
			return nil, xerrors.Errorf("Could not construct Tracer: %w", err)
		}
	}

	return tracer, nil
}

// buildErrChain builds a slice of all of the errors with the oldest at the back of the list
func buildErrorChain(baseErr error) []error {
	chain := []error{}
	errCursor := baseErr
	for errCursor != nil {
		chain = append(chain, errCursor)
		errCursor = xerrors.Unwrap(errCursor)
	}

	return chain
}

// Read implements the io.Reader interface. Will read up to len(dest) bytes of the current error.
// Note that this means dest will only be filled up the contents of the error, regardless of if there are other errors
// to be read in the error stack.
// Returns io.EOF when there are no more errors to read, but notably will not be returned when the last error is returned.
func (tracer *Tracer) Read(dest []byte) (n int, err error) {
	if tracer.buffer.Len() == 0 && len(tracer.errorChain) == 0 {
		return 0, io.EOF
	} else if tracer.buffer.Len() == 0 {
		message := generateErrorString(tracer.popChain(), tracer.formatter, tracer.detailedOutput)
		// If we are passed a zero length error, returning an io.EOF is not appropriate.
		if len(message) == 0 {
			message = emptyError
		}

		tracer.buffer.WriteString(message)
	}

	return tracer.buffer.Read(dest)
}

// ReadNext will read one unwrapped error and its associated trace
// If Read() has been called, but the buffer has not been exhausted, its contents will be discarded.
// Returns io.EOF when there are no more errors to read, but notably will not be returned when the last error is returned.
func (tracer *Tracer) ReadNext() (string, error) {
	tracer.buffer.Reset()
	if len(tracer.errorChain) == 0 {
		return "", io.EOF
	}

	message := generateErrorString(tracer.popChain(), tracer.formatter, tracer.detailedOutput)
	if len(message) == 0 {
		return emptyError, nil
	}

	return message, nil
}

// popChain will pop the next error off the error chain
func (tracer *Tracer) popChain() (storedError error) {
	nextErrorIndex := 0
	if tracer.ordering == OldestFirstOrdering {
		nextErrorIndex = len(tracer.errorChain) - 1
	}
	storedError = tracer.errorChain[nextErrorIndex]
	tracer.errorChain = tracer.errorChain[:nextErrorIndex]

	return
}

// Format allows for tracer to implement fmt.Formatter. This will simply make a clone of the formatter
// and print out the full trace. DetailedOutput will be given when %+v is provided, and normal output
// when %v is provided
func (tracer *Tracer) Format(s fmt.State, verb rune) {
	if verb != 'v' {
		return
	}

	clone, err := NewTracer(tracer.baseErr, tracer.optionFuncs...)
	if err != nil {
		out := fmt.Sprintf("<could not print trace: %s>", err)
		io.WriteString(s, out)
		return
	}

	clone.detailedOutput = s.Flag('+')
	err = nil
	outBuffer := bytes.NewBufferString("")
	for {
		var out string
		out, err = clone.ReadNext()
		// We don't even want to attempt to write to the buffer if there's an error
		if err != nil {
			break
		}

		outBuffer.WriteString(out + "\n")
	}

	if err != nil && err != io.EOF {
		out := fmt.Sprintf("<could not print trace: %s>", err)
		io.WriteString(s, out)
		return
	}

	// Strip out the final newline.
	outBuffer.Truncate(outBuffer.Len() - 1)
	io.Copy(s, outBuffer)
}
