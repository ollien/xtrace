package xtrace

/**
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

// Naive will set the naive flag when passed to NewNewLineFormatter. This flag, if set, will instruct the formatter
// to perform the naive version of this algorithm, which simply adds/removes a newline from the end of each message.
// xerrors has a habit of sending indentation in the previous line (i.e. "<error>\n    "), so the naive algorithm
// produces weird output. Nevertheless, this may be desirable depending on the implementation of xerrors.Formatter, so
// it is left as an option.
func Naive(naive bool) func(*NewLineFormatter) error {
	return func(formatter *NewLineFormatter) error {
		formatter.naive = naive

		return nil
	}
}

// NestingIndentation sets the indentation for the NestedMessageFormatter that is produced when this is passed to
// NewNestedMessageFormatter.
func NestingIndentation(indentation string) func(*NestedMessageFormatter) error {
	return func(formatter *NestedMessageFormatter) error {
		formatter.indentation = indentation

		return nil
	}
}
