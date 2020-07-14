package ss

import (
	"errors"
	"github.com/anujga/dstk/pkg/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type MsgHandler struct {
	w WorkerActor
}

func (mh *MsgHandler) Handle(req Msg) ([]interface{}, error) {
	select {
	case mh.w.Mailbox() <- req:
	default:
		return nil, core.ErrInfo(codes.ResourceExhausted, "Worker busy",
			"capacity", cap(mh.w.Mailbox())).Err()
	}
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

func (mh *MsgHandler) HandleBlocking(req Msg) (interface{}, *status.Status) {
	select {
	case mh.w.Mailbox() <- req:
	default:
		return nil, core.ErrInfo(codes.ResourceExhausted, "Worker busy",
			"capacity", cap(mh.w.Mailbox()))
	}

	select {
	case e := <-req.ResponseChannel():
		switch v := e.(type) {
		case error:
			return nil, status.Convert(v)
		default:
			return v, nil
		}

	case _ = <-time.After(time.Second * 5):
		return nil, core.ErrInfo(codes.DeadlineExceeded, "timeout",
			"msg", req)
	}
}
