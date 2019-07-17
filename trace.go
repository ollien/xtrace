package xtrace

/*
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
	"io"
	"os"

	"golang.org/x/xerrors"
)

// Trace prints a trace of errors wrapped by xerrors to stderr, with a terminating newline. If more customization is
// desired, please use Tracer.
func Trace(baseErr error) error {
	return traceToWriter(baseErr, os.Stderr)
}

// traceToWriter creates a Tracer and calls trace on it.
func traceToWriter(baseErr error, writer io.Writer) error {
	tracer, err := NewTracer(baseErr)
	if err != nil {
		return xerrors.Errorf("failed to initialize trace: %w", err)
	}

	err = tracer.trace(writer)
	if err != nil {
		return xerrors.Errorf("failed to run trace: %w", err)
	}

	// The default tracer does not end with a newline, so write one.
	_, err = writer.Write([]byte("\n"))
	if err != nil {
		return xerrors.Errorf("could not complete trace to stderr: %w", err)
	}

	return nil
}
