package handler

type Error struct{
	err error
}

func (e *Error) Error() string{
	return e.err.Error()
}