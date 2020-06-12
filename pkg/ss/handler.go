package ss

import (
	"errors"
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
		*e = elem.(error)
		*response = (*e).Error()
	default:
		*response = "internal error"
		*e = errors.New("invalid response")
	}
	return
}

type MsgHandler struct {
	Router
}

func (mh *MsgHandler) Handle(req Msg) (interface{}, error) {
	if err := mh.OnMsg(req); err != nil {
		// TODO find a better place to close this
		close(req.ResponseChannel())
		return "", err.Err()
	} else {
		var response interface{}
		var errToRet error
		for {
			select {
			case e, ok := <-req.ResponseChannel():
				if !ok {
					return response, errToRet
				}
				handleResponseElement(e, &response, &errToRet)
			case _ = <-time.After(time.Second * 5):
				return "internal error", errors.New("timeout")
			}
		}
	}
}
