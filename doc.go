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

/*
Package xtrace provides the ability to generate a trace of wrapped errors from xerrors. This is facilitated through the
Tracer type, the output of which can be customized with a TraceFormatter. For more information on how to wrap errors,
see https://godoc.org/golang.org/x/xerrors.

Basic Usage

The following example will print a trace of all of the wrapped errors to stderr.

	package main

	import (
		"errors"

		"github.com/ollien/xtrace"
		"golang.org/x/xerrors"
	)

	func main() {
		baseErr := errors.New("aw shucks, something broke")
		err2 := xerrors.Errorf("things went wrong!: %w", baseErr)

		traceErr := xtrace.Trace(err2)
		if traceErr != nil {
			panic("can not trace")
		}
		// aw shucks, something broke
		// things went wrong!
		// github.com/ollien/xtrace.ExampleTracer_Format
		//    /home/nick/Documents/code/xtrace/example.go:12
	}

If more customization is desired, one can use a Tracer. One of Tracer's key features is its compatibility with fmt.

	// ...
	tracer, err := xtrace.NewTracer(err2)
	if err != nil {
		panic("can not make tracer")
	}

	fmt.Printf("%v", tracer)
	// aw shucks, something broke
	// things went wrong!

You can also add %+v for more detailed information.
	// ...
	fmt.Printf("%+v", tracer)
	// aw shucks, something broke
	// things went wrong!
	// github.com/ollien/xtrace.ExampleTracer_Format
	//    /home/nick/Documents/code/xtrace/example.go:18

Using fmt is not required, though. You may instead read the errors one at a time from the trace with the ReadNext and
Read functions.

	// ...
	output, err := tracer.ReadNext()
	if err != nil {
		panic("can not read from tracer")
	}

	fmt.Println(output)
	// aw shucks, something broke

Customization

All output of a Tracer can be customized. By default, the Tracer will ensure that all messages end in a newline. If you
want more customization than that, then you can create your own TraceFormatter.
For instance, to make all of your errors in all caps, you can use the following TraceFormatter.

	type capsFormatter struct{}

	func (formatter capsFormatter) FormatTrace(previous []string, message string) string {
		return strings.ToUpper(message)
	}

You can then set a Tracer's TraceFormatter like so
	tracer, err := NewTracer(err, Formatter(capsFormatter{}))
*/
package xtrace
