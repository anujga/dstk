package control

import "google.golang.org/grpc/status"

type Response struct {
	Res interface{}
	Err *status.Status
}

func Success(r interface{}) *Response {
	return &Response{Res: r}
}

func Failure(err *status.Status) *Response {
	return &Response{Err: err}
}

func FailureErr(err error) *Response {
	return &Response{Err: status.Convert(err)}
}
