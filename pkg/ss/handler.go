package ss

import (
	"errors"
	"time"
)

type MsgHandler struct {
	Router
}

func (mh *MsgHandler) Handle(req Msg) ([]interface{}, error) {
	if err := mh.OnMsg(req); err != nil {
		// TODO find a better place to close this
		close(req.ResponseChannel())
		return nil, err
	} else {
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
}
