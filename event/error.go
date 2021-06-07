package event

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