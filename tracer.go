package xtrace

import (
	"bytes"

	"golang.org/x/xerrors"
)

// Tracer gets the trace of an error
type Tracer struct {
	baseErr, nextErr error
	detailedOutput   bool
	// Will only ever hold at most the contents of the current error being read
	buffer *bytes.Buffer
	// The number of bytes read from the buffer
	readBytes int
}

// NewTracer returns a new tracer for the given error
func NewTracer(err error) Tracer {
	detailedOutput := true
	message, nextErr := generateErrorString(err, detailedOutput)
	return Tracer{
		baseErr:        err,
		nextErr:        nextErr,
		detailedOutput: detailedOutput,
		buffer:         bytes.NewBufferString(message),
	}
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
func (tracer *Tracer) ReadNext() (message string, err error) {
	defer func() {
		if err != nil {
			return
		}

		err = tracer.shiftErrors()
	}()

	if tracer.buffer.Len() > 0 && tracer.readBytes > 0 {
		err = tracer.shiftErrors()
		if err != nil {
			return
		}
	}

	output := tracer.buffer.String()
	tracer.readBytes = len(output)

	return output, nil
}

// shiftErrors will get us the nextErr and hydrate the buffer with it
func (tracer *Tracer) shiftErrors() error {
	message, next := generateErrorString(tracer.nextErr, tracer.detailedOutput)
	tracer.nextErr = next
	tracer.buffer.Reset()
	_, err := tracer.buffer.WriteString(message)
	if err != nil {
		return xerrors.Errorf("can not write to error buffer: %w", err)
	}

	return nil
}
