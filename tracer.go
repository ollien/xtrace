package main

import (
	"bytes"

	"golang.org/x/xerrors"
)

// Tracer gets the trace of an error
type Tracer struct {
	baseErr, nextErr error
	// Will only ever hold at most the contents of the current error being read
	buffer *bytes.Buffer
}

// NewTracer returns a new tracer for the given error
func NewTracer(err error) Tracer {
	message := err.Error()
	nextErr := xerrors.Unwrap(err)
	return Tracer{
		baseErr: err,
		nextErr: nextErr,
		buffer:  bytes.NewBufferString(message),
	}
}

// Read implements the io.Reader interface. Will read up to len(dest) bytes of the current error.
// Note that this means dest will only be filled up the contents of the error, regardless of if there are other errors
// to be read in the error stack.
func (tracer *Tracer) Read(dest []byte) (n int, err error) {
	// TODO: Implement
	return 0, nil
}

// ReadNext will read one unwrapped error and its associated trace
// If Read() has been called, but the buffer has not been exhausted, its contents will be discarded.
func (tracer *Tracer) ReadNext() {
	// TODO: Implement
}
