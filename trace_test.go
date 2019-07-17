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
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"
)

type traceTest struct {
	name     string
	testFunc func(*testing.T)
}

func runTraceTestTable(t *testing.T, table []traceTest) {
	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestTrace(t *testing.T) {
	tests := []traceTest{
		traceTest{
			name: "basic error",
			testFunc: func(t *testing.T) {
				buffer := bytes.NewBufferString("")
				err := errors.New("things broke :(")
				traceErr := traceToWriter(err, buffer)
				assert.Nil(t, traceErr)
				assert.Equal(t, err.Error()+"\n", buffer.String())
			},
		},
		traceTest{
			name: "wrapped errors",
			testFunc: func(t *testing.T) {
				buffer := bytes.NewBufferString("")
				err := errors.New("things broke :(")
				err2 := xerrors.Errorf("aw shucks: %w", err)
				traceErr := traceToWriter(err2, buffer)
				assert.Nil(t, traceErr)

				bufferString := buffer.String()
				// There may be more error contents given the fact that detailed output is out by default. Just ensure
				// we have the contents that we expect
				assert.Equal(t, 1, strings.Count(bufferString, "things broke :("))
				assert.Equal(t, 1, strings.Count(bufferString, "aw shucks"))
				assert.Equal(t, byte('\n'), bufferString[len(bufferString)-1])
				assert.True(t, func() bool {
					return strings.Index(bufferString, "things broke :(") < strings.Index(bufferString, "aw shucks")
				}())
			},
		},
	}

	runTraceTestTable(t, tests)
}
