package xtrace

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"
)

type test struct {
	name     string
	setup    func(t *testing.T) Tracer
	testFunc func(t *testing.T, tracer Tracer)
}

func runTestTable(t *testing.T, table []test) {
	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			tracer := tt.setup(t)
			tt.testFunc(t, tracer)
		})
	}
}

func handleSetupError(t *testing.T, tracer Tracer, err error) Tracer {
	if err != nil {
		t.Log("Could not setup test", err)
		t.FailNow()

		// Won't ever happen after FailNow
		return Tracer{}
	}

	return tracer
}

func TestTracer_ReadNext(t *testing.T) {
	tests := []test{
		test{
			name: "two nested errors",
			setup: func(t *testing.T) Tracer {
				err := errors.New("things broke :(")
				err2 := xerrors.Errorf("aw shucks: %w", err)
				err3 := xerrors.Errorf("I tried very hard and failed: %w", err2)

				tracer, constructErr := NewTracer(err3)

				return handleSetupError(t, tracer, constructErr)
			},
			testFunc: func(t *testing.T, tracer Tracer) {
				var err error
				expectedErrors := []string{
					"things broke :(",
					"aw shucks",
					"I tried very hard and failed",
				}
				for i := 0; err != io.EOF; i++ {
					if i >= len(expectedErrors)+1 {
						fmt.Printf("Ran more times than expected: (on attempt %d, only expected %d)", i+1, len(expectedErrors))
						t.FailNow()
					}

					var message string
					message, err = tracer.ReadNext()
					if err == nil {
						// Other details may be returned when we use a tracer, so we only want to assert that the expected message is there
						assert.Equal(t, 1, strings.Count(message, expectedErrors[i]))
						// Make sure that the next error is not contained in our current message
						if i != len(expectedErrors)-1 {
							assert.NotContains(t, message, expectedErrors[i+1])
						}
					} else {
						assert.Equal(t, io.EOF, err)
					}
				}
			},
		},
		test{
			name: "nil error",
			setup: func(t *testing.T) Tracer {
				tracer, constructErr := NewTracer(nil)

				return handleSetupError(t, tracer, constructErr)
			},
			testFunc: func(t *testing.T, tracer Tracer) {
				_, err := tracer.ReadNext()
				assert.Equal(t, io.EOF, err)
			},
		},
		test{
			name: "empty error",
			setup: func(t *testing.T) Tracer {
				err := errors.New("")
				tracer, constructErr := NewTracer(err)

				return handleSetupError(t, tracer, constructErr)
			},
			testFunc: func(t *testing.T, tracer Tracer) {
				message, err := tracer.ReadNext()
				assert.Equal(t, message, emptyError)
				assert.Nil(t, err)
			},
		},
		test{
			name: "reset Read",
			setup: func(t *testing.T) Tracer {
				err := errors.New("things broke :(")
				err2 := xerrors.Errorf("aw shucks: %w", err)
				err3 := xerrors.Errorf("I tried very hard and failed: %w", err2)
				tracer, constructErr := NewTracer(err3)

				return handleSetupError(t, tracer, constructErr)
			},
			testFunc: func(t *testing.T, tracer Tracer) {
				buffer := make([]byte, 5)
				n, err := tracer.Read(buffer)
				assert.Nil(t, err)
				assert.Equal(t, 5, n)
				assert.Equal(t, "thing", string(buffer))
				// Make sure the next error we read is not "things broke :("
				message, err := tracer.ReadNext()
				assert.Nil(t, err)
				assert.Equal(t, 1, strings.Count(message, "aw shucks"))

				// Make sure the next call to Read does not pick up where it left off
				n, err = tracer.Read(buffer)
				assert.Nil(t, err)
				assert.Equal(t, 5, n)
				assert.Equal(t, "I tri", string(buffer))
			},
		},
	}

	runTestTable(t, tests)
}

func TestTracer_Read(t *testing.T) {
	tests := []test{
		test{
			name: "no errors",
			setup: func(t *testing.T) Tracer {
				tracer, constructErr := NewTracer(nil)

				return handleSetupError(t, tracer, constructErr)
			},
			testFunc: func(t *testing.T, tracer Tracer) {
				buffer := make([]byte, 15)
				n, err := tracer.Read(buffer)
				assert.Equal(t, 0, n)
				assert.Equal(t, err, io.EOF)
				assert.Equal(t, make([]byte, 15), buffer)
			},
		},
		test{
			name: "one error, full read",
			setup: func(t *testing.T) Tracer {
				err := errors.New("things broke :(")
				tracer, constructErr := NewTracer(err)

				return handleSetupError(t, tracer, constructErr)
			},
			testFunc: func(t *testing.T, tracer Tracer) {
				buffer := make([]byte, len("things broke :("))
				n, err := tracer.Read(buffer)
				assert.Equal(t, len(buffer), n)
				assert.Nil(t, err)
				assert.Equal(t, "things broke :(", string(buffer))

				n, err = tracer.Read(buffer)
				assert.Equal(t, 0, n)
				assert.Equal(t, err, io.EOF)
				assert.Equal(t, "things broke :(", string(buffer))
			},
		},
		test{
			name: "one error, many reads",
			setup: func(t *testing.T) Tracer {
				err := errors.New("things broke :(")
				tracer, constructErr := NewTracer(err)

				return handleSetupError(t, tracer, constructErr)
			},
			testFunc: func(t *testing.T, tracer Tracer) {
				buffer := make([]byte, 5)
				fullBuffer := make([]byte, 0)
				totalN := 0
				n, err := 0, error(nil)
				for {
					n, err = tracer.Read(buffer)
					totalN += n
					if err == io.EOF {
						break
					}

					fullBuffer = append(fullBuffer, buffer...)
					assert.Nil(t, err)
					assert.True(t, func() bool {
						return n <= len(buffer) && n > 0
					}())
				}

				assert.Equal(t, len(fullBuffer), totalN)
				assert.Equal(t, 0, n)
				assert.Equal(t, "things broke :(", string(fullBuffer))

				n, err = tracer.Read(buffer)
				assert.Equal(t, n, 0)
				assert.Equal(t, io.EOF, err)
				assert.Equal(t, "things broke :(", string(fullBuffer))
			},
		},
		test{
			name: "many errors, one read",
			setup: func(t *testing.T) Tracer {
				err := errors.New("things broke :(")
				err2 := xerrors.Errorf("aw shucks: %w", err)
				err3 := xerrors.Errorf("I tried very hard and failed: %w", err2)
				tracer, constructErr := NewTracer(err3)

				return handleSetupError(t, tracer, constructErr)
			},
			testFunc: func(t *testing.T, tracer Tracer) {
				buffer := make([]byte, len("things broke :(")*2)
				n, err := tracer.Read(buffer)
				assert.Equal(t, len(buffer)/2, n)
				assert.Nil(t, err)
				// No matter our buffer size, we only want to get the first error back
				// Even though there are many errors, because we are only reading the first one, adn that one is just
				// a simple error, we don't have to worry about there being contents other than the error message.
				expectedBuffer := make([]byte, len(buffer))
				for i, char := range "things broke :(" {
					expectedBuffer[i] = byte(char)
				}
				assert.Equal(t, string(expectedBuffer), string(buffer))
			},
		},
		test{
			name: "many errors, many reads",
			setup: func(t *testing.T) Tracer {
				err := errors.New("things broke :(")
				err2 := xerrors.Errorf("aw shucks: %w", err)
				err3 := xerrors.Errorf("I tried very hard and failed: %w", err2)
				tracer, constructErr := NewTracer(err3)

				return handleSetupError(t, tracer, constructErr)
			},
			testFunc: func(t *testing.T, tracer Tracer) {
				expectedErrors := []string{
					"things broke :(",
					"aw shucks",
					"I tried very hard and failed",
				}
				buffer := make([]byte, 5)
				fullBuffer := make([]byte, 0)
				totalN := 0
				n, err := 0, error(nil)
				for {
					n, err = tracer.Read(buffer)
					totalN += n
					if err == io.EOF {
						break
					}

					fullBuffer = append(fullBuffer, buffer[:n]...)
					assert.Nil(t, err)
					assert.True(t, func() bool {
						return n <= len(buffer) && n > 0
					}())
				}

				assert.Equal(t, len(fullBuffer), totalN)
				assert.Equal(t, 0, n)
				for _, expectedError := range expectedErrors {
					assert.Equal(t, 1, bytes.Count(fullBuffer, []byte(expectedError)))
				}

				fullBufferClone := make([]byte, len(fullBuffer))
				copy(fullBufferClone, fullBuffer)
				n, err = tracer.Read(buffer)
				assert.Equal(t, n, 0)
				assert.Equal(t, io.EOF, err)
				assert.Equal(t, string(fullBufferClone), string(fullBuffer))
			},
		},
		test{
			name: "many errors, test error boundary",
			setup: func(t *testing.T) Tracer {
				err := errors.New("things broke :(")
				err2 := xerrors.Errorf("aw shucks: %w", err)
				tracer, constructErr := NewTracer(err2)

				return handleSetupError(t, tracer, constructErr)
			},
			testFunc: func(t *testing.T, tracer Tracer) {
				buffer := make([]byte, 6)
				fullBuffer := make([]byte, 0, 18)
				// Exhaust the buffer, and ensure we don't hit an io.EOF
				for i := 0; i < 3; i++ {
					n, err := tracer.Read(buffer)
					assert.True(t, func() bool {
						return n > 0
					}())
					assert.Nil(t, err)
					for j := 0; j < n; j++ {
						fullBuffer = append(fullBuffer, buffer[j])
					}
				}
				// Ensure that we ONLY have the expected error
				assert.Equal(t, "things broke :(", string(bytes.TrimRight(fullBuffer, "\x00")))
			},
		},
		test{
			name: "empty error",
			setup: func(t *testing.T) Tracer {
				err := errors.New("")
				tracer, constructErr := NewTracer(err)

				return handleSetupError(t, tracer, constructErr)
			},
			testFunc: func(t *testing.T, tracer Tracer) {
				buffer := make([]byte, len(emptyError))
				n, err := tracer.Read(buffer)
				assert.Equal(t, len(buffer), n)
				assert.Nil(t, err)
				assert.Equal(t, emptyError, string(buffer))

				n, err = tracer.Read(buffer)
				assert.Equal(t, 0, n)
				assert.Equal(t, err, io.EOF)
				assert.Equal(t, emptyError, string(buffer))
			},
		},
	}

	runTestTable(t, tests)
}

type capsFormatter struct{}

func (formatter capsFormatter) FormatTrace(previous []string, message string) string {
	return strings.ToUpper(message)
}

func ExampleNewTracer() {
	baseErr := errors.New("aw shucks, something broke")
	// capsFormatter is a custom formatter that simply applies strings.ToUpper to all messages
	tracer, err := NewTracer(baseErr, Formatter(capsFormatter{}))
	if err != nil {
		panic("can not make tracer")
	}
	output, err := tracer.ReadNext()
	if err != nil {
		panic("can not read from tracer")
	}

	fmt.Println(output)
	// Output: AW SHUCKS, SOMETHING BROKE
}
