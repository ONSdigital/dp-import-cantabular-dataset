package handler

// Error is the handler package's error type. Is not meant to be compared as
// a type, but information should be extracted via the interfaces
// it implements with callback functions. Is not guaranteed to remain exported
// so shouldn't be treated as such.
type Error struct {
	err               error
	logData           map[string]interface{}
	instanceCompleted bool
}

// NewError creates a new Error
func NewError(err error, logData map[string]interface{}, instanceCompleted bool) *Error {
	return &Error{
		err:               err,
		logData:           logData,
		instanceCompleted: instanceCompleted,
	}
}

// Error implements the Go standard error interface
func (e *Error) Error() string {
	return e.err.Error()
}

// LogData implements the DataLogger interface which allows you extract
// embedded log.Data from an error
func (e *Error) LogData() map[string]interface{} {
	return e.logData
}

// Unwrap implements Go's error wrapping interface
func (e *Error) Unwrap() error {
	return e.err
}

// InstanceCompleted implements the instanceCompleteder interface and allows
// you to extract the instanceCompleted flag from an error. Is specific to this
// application and unlikely to be found elsewhere
func (e *Error) InstanceCompleted() bool {
	return e.instanceCompleted
}
