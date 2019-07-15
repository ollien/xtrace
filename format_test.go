package xtrace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type formatTest struct {
	name     string
	setup    func(t *testing.T) TraceFormatter
	testFunc func(t *testing.T, formatter TraceFormatter)
}

func runFormatTestTable(t *testing.T, table []formatTest) {
	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			tracer := tt.setup(t)
			tt.testFunc(t, tracer)
		})
	}
}

func handleFormatTestSetupError(t *testing.T, formatter TraceFormatter, err error) TraceFormatter {
	if err != nil {
		t.Log("Could not setup test", err)
		t.FailNow()

		// Won't ever happen after FailNow
		return nil
	}

	return formatter
}

func TestNilFormatter(t *testing.T) {
	tests := []formatTest{
		formatTest{
			name: "one error",
			setup: func(t *testing.T) TraceFormatter {
				return NilFormatter{}
			},
			testFunc: func(t *testing.T, formatter TraceFormatter) {
				output := formatter.FormatTrace(nil, "    hello   \n")
				assert.Equal(t, "    hello   \n", output)
			},
		},
		formatTest{
			name: "many errors",
			setup: func(t *testing.T) TraceFormatter {
				return NilFormatter{}
			},
			testFunc: func(t *testing.T, formatter TraceFormatter) {
				trace := []string{}
				messages := []string{
					"things broke :(",
					"an awful thing happened",
					"aw shucks",
				}
				for _, message := range messages {
					formattedOutput := formatter.FormatTrace(trace, message)
					assert.Equal(t, message, formattedOutput)
					trace = append(trace, formattedOutput)
				}

				assert.Equal(t, []string{"things broke :(", "an awful thing happened", "aw shucks"}, trace)
			},
		},
	}

	runFormatTestTable(t, tests)
}

func TestNewLineFormatter(t *testing.T) {
	tests := []formatTest{
		formatTest{
			name: "one error, non-naive",
			setup: func(t *testing.T) TraceFormatter {
				formatter, err := NewNewLineFormatter()

				return handleFormatTestSetupError(t, formatter, err)
			},
			testFunc: func(t *testing.T, formatter TraceFormatter) {
				output := formatter.FormatTrace(nil, "    hello   \n")
				assert.Equal(t, "    hello   ", output)
			},
		},
		formatTest{
			name: "one error, non-naive and mid-error newline",
			setup: func(t *testing.T) TraceFormatter {
				formatter, err := NewNewLineFormatter()

				return handleFormatTestSetupError(t, formatter, err)
			},
			testFunc: func(t *testing.T, formatter TraceFormatter) {
				output := formatter.FormatTrace(nil, "    hello\n   \n")
				assert.Equal(t, "    hello   ", output)
			},
		},
		formatTest{
			name: "many errors, non-naive",
			setup: func(t *testing.T) TraceFormatter {
				formatter, err := NewNewLineFormatter()

				return handleFormatTestSetupError(t, formatter, err)
			},
			testFunc: func(t *testing.T, formatter TraceFormatter) {
				trace := []string{}
				messages := []string{
					"things broke :(",
					"an awful thing happened",
					"aw shucks",
				}
				for _, message := range messages {
					formattedOutput := formatter.FormatTrace(trace, message)
					// Each message should be the last message at the time of insertion, so none of them should have new lines
					assert.Equal(t, message, formattedOutput)
					trace = append(trace, formattedOutput)
				}

				assert.Equal(t, []string{"things broke :(\n", "an awful thing happened\n", "aw shucks"}, trace)
			},
		},
		formatTest{
			name: "one error, naive",
			setup: func(t *testing.T) TraceFormatter {
				formatter, err := NewNewLineFormatter(Naive(true))

				return handleFormatTestSetupError(t, formatter, err)
			},
			testFunc: func(t *testing.T, formatter TraceFormatter) {
				output := formatter.FormatTrace(nil, "    hello   \n")
				assert.Equal(t, "    hello   ", output)
			},
		},
		formatTest{
			name: "one error, naive and mid-error newline",
			setup: func(t *testing.T) TraceFormatter {
				formatter, err := NewNewLineFormatter(Naive(true))

				return handleFormatTestSetupError(t, formatter, err)
			},
			testFunc: func(t *testing.T, formatter TraceFormatter) {
				output := formatter.FormatTrace(nil, "    hello\n   ")
				assert.Equal(t, "    hello\n   ", output)
			},
		},
		formatTest{
			name: "many errors, naive",
			setup: func(t *testing.T) TraceFormatter {
				formatter, err := NewNewLineFormatter(Naive(true))

				return handleFormatTestSetupError(t, formatter, err)
			},
			testFunc: func(t *testing.T, formatter TraceFormatter) {
				trace := []string{}
				messages := []string{
					"things broke :(",
					"an awful thing happened",
					"aw shucks",
				}
				for _, message := range messages {
					formattedOutput := formatter.FormatTrace(trace, message)
					// Each message should be the last message at the time of insertion, so none of them should have new lines
					assert.Equal(t, message, formattedOutput)
					trace = append(trace, formattedOutput)
				}

				assert.Equal(t, []string{"things broke :(\n", "an awful thing happened\n", "aw shucks"}, trace)
			},
		},
	}

	runFormatTestTable(t, tests)
}

func TestNestedMessageFormatter(t *testing.T) {
	tests := []formatTest{
		formatTest{
			name: "one error",
			setup: func(t *testing.T) TraceFormatter {
				formatter, err := NewNestedMessageFormatter(NestingIndentation("\t"))

				return handleFormatTestSetupError(t, formatter, err)
			},
			testFunc: func(t *testing.T, formatter TraceFormatter) {
				output := formatter.FormatTrace(nil, "    hello   \n")
				assert.Equal(t, "hello", output)
			},
		},
		formatTest{
			name: "many errors, non-naive",
			setup: func(t *testing.T) TraceFormatter {
				formatter, err := NewNestedMessageFormatter(NestingIndentation("  "))

				return handleFormatTestSetupError(t, formatter, err)
			},
			testFunc: func(t *testing.T, formatter TraceFormatter) {
				trace := []string{}
				messages := []string{
					"things broke :(",
					"an awful thing happened",
					"aw shucks",
				}
				for i, message := range messages {
					formattedOutput := formatter.FormatTrace(trace, message)
					expected := message
					if i != 0 {
						expected = "  " + expected
					}
					// Each message should be the last message at the time of insertion, so none of them should have new lines
					assert.Equal(t, expected, formattedOutput)
					trace = append(trace, formattedOutput)
				}

				assert.Equal(t, []string{"things broke :(\n", "  an awful thing happened\n", "  aw shucks"}, trace)
			},
		},
	}

	runFormatTestTable(t, tests)
}
