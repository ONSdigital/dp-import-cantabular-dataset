package event

import (
	"errors"
)

type Error struct{
	err error
	logData map[string]interface{}
}

func (e *Error) Error() string{
	return e.err.Error()
}

func (e *Error) LogData() map[string]interface{}{
	return e.logData
}

// statusCode returns a statusCode from an error if there is one
func statusCode(err error) int{
	var cerr coder
	if errors.As(err, &cerr){
		return cerr.Code()
	}

	return 0
}

// logData returns logData for an error if there is any
func logData(err error) map[string]interface{}{
	var lderr dataLogger
	if errors.As(err, &lderr){
		return lderr.LogData()
	}

	return nil
}

// unwrapLogData recursively unwraps logData from an error
func unwrapLogData(err error) []map[string]interface{}{
	var data []map[string]interface{}

	for errors.Unwrap(err) != nil {
		if newErr := errors.Unwrap(err); newErr != nil{
			if d := logData(err); d != nil{
				data = append(data, d)
			}
			err = newErr
		}
	}

	if d := logData(err); d != nil{
		data = append(data, d)
	}

	return data
}
