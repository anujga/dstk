package main

import (
	"errors"
	"github.com/anujga/dstk/pkg/ss"
	"time"
)

func handleResponseElement(elem interface{}, response *string, e *error) {
	switch elem.(type) {
	case string:
		*response = elem.(string)
		*e = nil
	case error:
		err := elem.(error)
		*response = err.Error()
		*e = err
	default:
		*response = "internal error"
		*e = errors.New("invalid response")
	}
	return
}

type ReqHandler struct {
	router ss.Router
}

//todo: this should happen in `PartitionMgr::OnMsg`
func (rh *ReqHandler) handle(req *Request) (string, error) {
	if err := rh.router.OnMsg(req); err != nil {
		// TODO find a better place to close this
		close(req.C)
		return "", err
	} else {
		var response string
		var errToRet error
		for {
			select {
			case e, ok := <-req.C:
				if !ok {
					return response, errToRet
				}
				handleResponseElement(e, &response, &errToRet)
			case _ = <-time.After(time.Second * 5):
				return "internal error", errors.New("timedout")
			}
		}
	}
}
