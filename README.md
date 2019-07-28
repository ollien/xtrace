# xtrace

[![](https://godoc.org/github.com/ollien/xtrace?status.svg)](http://godoc.org/github.com/ollien/xtrace)
[![Build Status](https://travis-ci.com/ollien/xtrace.svg?branch=master)](https://travis-ci.com/ollien/xtrace)

[xerrors](https://godoc.org/golang.org/x/xerrors) is pretty awesome. It allows you to wrap your errors and provide more context. Sadly, one of the features of it that isn't really available until Go 1.13 is the ability to print out a stack trace. No longer!

## Installation

`go get -u github.com/ollien/xtrace`

## Basic Usage

The following example will print a trace of all of the wrapped errors to stderr.
```go
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
```

If more customization is desired, one can use a Tracer. One of Tracer's key features is its compatibility with fmt.

```go
// ...
tracer, err := xtrace.NewTracer(err2)
if err != nil {
	panic("can not make tracer")
}

fmt.Printf("%v", tracer)
// aw shucks, something broke
// things went wrong!
```

That's nice, and we can see a trace of all of our errors, but xerrors provides much more for information for us. To get this information, all we have to do is change our format string to %+v.

```go
// ... (repeated from above)

fmt.Printf("%+v", tracer)
// aw shucks, something broke
// things went wrong!
// github.com/ollien/xtrace.ExampleTracer_Format
//    /home/nick/Documents/code/xtrace/example.go:18
```

If you don't want to use `fmt` and just want to get this trace as a string, you can simply use `ReadNext` to get the next error in the trace. (Tracer also implements `io.Reader` if you prefer to use that.)

```go
// ... (repeated from above)

output, err := tracer.ReadNext()
if err != nil {
	panic("can not read from tracer")
}

fmt.Println(output)
// aw shucks, something broke
```

See the docs for more usages.

## Customization

Output customization is one of the explicit goals of xtrace. For instance, if you wish to flip the output of your trace so that the newest errors are on top (i.e. with the root cause at the bottom), all you have to do is the following.
```go
tracer, err := NewTracer(err, Ordering(NewestFirstOrdering))
```

You can also set up custom formatters for your traces. There are several included (such as `NestedMessageFormatter` and `NilFormatter`). If you want, you can also write your own formatter by simply implementing the `TraceFormatter` interface. For instance, if you wanted to make sure that _everyone_ hears your errors, you can make all of them capitalized.

```go
type capsFormatter struct{}

func (formatter capsFormatter) FormatTrace(previous []string, message string) string {
	return strings.ToUpper(message)
}
```

You can then set a Tracer's formatter by doing
```go
tracer, err := NewTracer(err, Formatter(capsFormatter{}))
```

If you wish to combine your loud errors with newest-first ordering you can pass them both as arguments to `NewTracer`.

```go
tracer, err := NewTracer(err, Ordering(NewestFirstOrdering), Formatter(capsFormatter{}))
```

See the docs for more details.