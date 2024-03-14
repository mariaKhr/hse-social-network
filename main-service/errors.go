package main

type HttpError struct {
	Msg    string
	Status int
}

func (e *HttpError) Error() string {
	return e.Msg
}

func (a *HttpError) HttpStatus() int {
	return a.Status
}
