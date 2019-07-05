package xtrace

import (
	"bytes"
	"io"

	"golang.org/x/xerrors"
)

// Tracer gets the trace of an error
type Tracer struct {
	detailedOutput bool
	// Populated with the full chain of errors, with the originating error at len(errorChain) - 1
	errorChain []error
	// Holds the contents of the current error being read
	buffer *bytes.Buffer
}

// NewTracer returns a new tracer for the given error
func NewTracer(err error) Tracer {
	detailedOutput := true
	return Tracer{
		errorChain:     buildErrorChain(err),
		detailedOutput: detailedOutput,
		buffer:         nil,
	}
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
// Returns io.EOF when there are no more errors to read.
func (tracer *Tracer) Read(dest []byte) (n int, err error) {
	// TODO: Implement
	return 0, nil
}

// ReadNext will read one unwrapped error and its associated trace
// If Read() has been called, but the buffer has not been exhausted, its contents will be discarded.
// Returns io.EOF when there are no more errors to read.
func (tracer *Tracer) ReadNext() (string, error) {
	if len(tracer.errorChain) == 0 {
		return "", io.EOF
	}

	return generateErrorString(tracer.popChain(), tracer.detailedOutput), nil
}

func (tracer *Tracer) popChain() (err error) {
	err, tracer.errorChain = tracer.errorChain[len(tracer.errorChain)-1], tracer.errorChain[:len(tracer.errorChain)-1]

	return
}
