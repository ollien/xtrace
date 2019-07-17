# xtrace

[![](https://godoc.org/github.com/ollien/xtrace?status.svg)](http://godoc.org/github.com/ollien/xtrace)

[xtrace](https://godoc.org/golang.org/x/xerrors) is pretty awesome. It allows you to wrap your errors and provide more context. Sadly, one of the features of it that isn't really available until Go 1.13 is the ability to print out a stack trace. No longer!

## Basic Usage

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
	tracer, err := xtrace.NewTracer(err2)
	if err != nil {
		panic("can not make tracer")
	}

	fmt.Printf("%v", tracer)
	// aw shucks, something broke
	// things went wrong!
}
```

That's nice, and we can see a trace of all of our errors, but xerrors provides much more for information for us. To get this information, all we have to do is change our format string to %+v.

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
	tracer, err := xtrace.NewTracer(err2)
	if err != nil {
		panic("can not make tracer")
	}

	fmt.Printf("%+v", tracer)
	// aw shucks, something broke
	// things went wrong!
	// github.com/ollien/xtrace.ExampleTracer_Format
	//    /home/nick/Documents/code/xtrace/example.go:18
}
```

If you don't want to use `fmt`, and just want to get this trace as a string, you can simply use `ReadNext` to get the next error in the trace. (Tracer also implements `io.Reader` if you prefer to use that.)

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
	tracer, err := xtrace.NewTracer(err2)
	if err != nil {
		panic("can not make tracer")
	}

	output, err := tracer.ReadNext()
	if err != nil {
		panic("can not read from tracer")
	}

	fmt.Println(output)
	// aw shucks, something broke
}
```

## Customization

Output customization is one of the explicit goals of xtrace. For instance, if you wish to flip the output of your trace, all you have to do is the following.
```go
tracer, err := NewTracer(err2, Ordering(NewestFirstOrdering))
```

You can also set up custom formatters for your traces. There are several included (such as NestedMessageFormatter and NilFormatter). If you want, you can also write your own formatter by simply implementing the `TraceFormatter` interface. For instance, if you wanted to make sure that _everyone_ hears your errors, you can make all of them capitalized.

```go
type capsFormatter struct{}

func (formatter capsFormatter) FormatTrace(previous []string, message string) string {
	return strings.ToUpper(message)
}
```

You can then set a tracer's formatter by doing
```go
tracer, err := NewTracer(err, Formatter(capsFormatter{}))
```


See the docs for more details.