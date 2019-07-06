package xtrace

// DetailedOutput will enable detailed output when passed to NewTracer. This detailed output is defined by the
// xerrors.Formatter for the passed error. Defaults to true.
func DetailedOutput(enabled bool) func(*Tracer) error {
	return func(tracer *Tracer) error {
		tracer.detailedOutput = enabled
		return nil
	}
}
