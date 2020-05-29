package main

import (
	"errors"
	"github.com/anujga/dstk/pkg/ss"
	"time"
)

func handleResponseElement(elem interface{}, response *string) {
	switch elem.(type) {
	case string:
		*response = elem.(string)
	case error:
		err := elem.(error)
		*response = err.Error()
	default:
		*response = "internal error"
	}
	return
}

type ReqHandler struct {
	router ss.Router
}

func (rh *ReqHandler) handle(req *Request) (string, error) {
	if err := rh.router.OnMsg(req); err != nil {
		return "", err
	} else {
		var response string
		for {
			select {
			case e, ok := <-req.C:
				if !ok {
					return response, nil
				}
				handleResponseElement(e, &response)
			case _ = <-time.After(time.Second * 5):
				return "internal error", errors.New("timedout")
			}
		}
	}
}
