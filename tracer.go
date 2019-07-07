package xtrace

import (
	"bytes"
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
}

// NewTracer returns a new tracer for the given error
func NewTracer(err error, options ...func(*Tracer) error) (Tracer, error) {
	tracer := Tracer{
		errorChain:     buildErrorChain(err),
		detailedOutput: true,
		buffer:         bytes.NewBuffer([]byte{}),
		formatter:      &NewLineFormatter{},
	}

	for _, optionFunc := range options {
		err := optionFunc(&tracer)
		if err != nil {
			return Tracer{}, xerrors.Errorf("Could not construct Tracer: %w", err)
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

func (tracer *Tracer) popChain() (err error) {
	err, tracer.errorChain = tracer.errorChain[len(tracer.errorChain)-1], tracer.errorChain[:len(tracer.errorChain)-1]

	return
}
