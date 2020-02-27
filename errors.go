package benchparse

type errorString string

func (e errorString) Error() string {
	return string(e)
}

// ErrNotMeasured indicates that a specific output
// was not measured.
const ErrNotMeasured = errorString("not measured")
