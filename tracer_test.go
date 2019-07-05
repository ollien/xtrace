package xtrace

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"
)

func TestTracer_ReadNext(t *testing.T) {
	type test struct {
		name     string
		setup    func() Tracer
		testFunc func(tracer Tracer)
	}
	tests := []test{
		test{
			name: "two nested errors",
			setup: func() Tracer {
				err := errors.New("things broke :(")
				err2 := xerrors.Errorf("aw shucks: %w", err)
				err3 := xerrors.Errorf("I tried very hard and failed: %w", err2)

				return NewTracer(err3)
			},
			testFunc: func(tracer Tracer) {
				var err error
				// Other details may be returned when we use a tracer, so we only want to assert that the expected message is at the start
				expectedPatterns := []string{
					`^things broke :\(`,
					`^aw shucks`,
					`^I tried very hard and failed`,
				}
				for i := 0; err != io.EOF; i++ {
					if i >= len(expectedPatterns)+1 {
						fmt.Printf("Ran more times than expected: (on attempt %d, only expected %d)", i+1, len(expectedPatterns))
						t.FailNow()
					}

					var message string
					message, err = tracer.ReadNext()
					if err == nil {
						assert.Regexp(t, expectedPatterns[i], message)
					} else {
						assert.Equal(t, io.EOF, err)
					}
				}
			},
		},
		test{
			name: "nil error",
			setup: func() Tracer {
				return NewTracer(nil)
			},
			testFunc: func(tracer Tracer) {
				_, err := tracer.ReadNext()
				assert.Equal(t, io.EOF, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracer := tt.setup()
			tt.testFunc(tracer)
		})
	}
}
