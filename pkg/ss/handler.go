package ss

import (
	"errors"
	"time"
)

type MsgHandler struct {
	WorkerActor
}

func (mh *MsgHandler) Handle(req Msg) ([]interface{}, error) {
	mh.Mailbox() <- req
	responses := make([]interface{}, 0)
	for {
		select {
		case e, ok := <-req.ResponseChannel():
			if ok {
				responses = append(responses, e)
			} else {
				return responses, nil
			}
		case _ = <-time.After(time.Second * 5):
			return nil, errors.New("timeout")
		}
	}
}
