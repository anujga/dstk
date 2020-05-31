package main

import (
	"errors"
	"github.com/anujga/dstk/pkg/ss"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func handleResponseElement(elem interface{}, response *interface{}, e *error) {
	switch elem.(type) {
	case int64:
		*response = elem.(int64)
		*e = nil
	case string:
		*response = elem.(string)
		*e = nil
	case error:
		err := elem.(error)
		*response = err.Error()
		*e = status.Errorf(codes.Internal, err.Error())
	default:
		*response = "internal error"
		*e = errors.New("invalid response")
	}
	return
}

type ReqHandler struct {
	router ss.Router
}

func (rh *ReqHandler) handle(req *Request) (interface{}, error) {
	if err := rh.router.OnMsg(req); err != nil {
		// TODO find a better place to close this
		close(req.C)
		return "", err
	} else {
		var response interface{}
		var errToRet error
		for {
			select {
			case e, ok := <-req.C:
				if !ok {
					return response, errToRet
				}
				handleResponseElement(e, &response, &errToRet)
			case _ = <-time.After(time.Second * 5):
				return "internal error", status.Errorf(codes.Aborted, "timeout")
			}
		}
	}
}
