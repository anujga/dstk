package main

import (
	"errors"
	"github.com/anujga/dstk/pkg/ss"
	"time"
)

type ReqHandler struct {
	router ss.Router
}

func (rh *ReqHandler) handle(req *Request) (string, error) {
	if err := rh.router.OnMsg(req); err != nil {
		return "", err
	} else {
		select {
		case e, ok := <-req.C:
			if !ok {
				return "ok", nil
			}
			switch e.(type) {
			case error:
				err = e.(error)
				return "internal error", err
			default:
				return "internal error", errors.New("invalid response")
			}
		case _ = <-time.After(time.Second * 5):
			return "internal error", errors.New("timedout")
		}
	}
}
