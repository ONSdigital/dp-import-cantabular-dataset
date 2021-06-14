package handler

type Error struct{
	err error
	statusCode int
	logData map[string]interface{}
}

func (e *Error) Error() string{
	return e.err.Error()
}

func (e *Error) Code() int{
	return e.statusCode
}

func (e *Error) LogData() map[string]interface{}{
	return e.logData
}

func (e *Error) Unwrap() error{
	return e.err
}