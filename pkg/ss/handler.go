package ss

import (
	"errors"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/common"
	"github.com/anujga/dstk/pkg/ss/node"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type MsgHandler struct {
	w node.Actor
}

func (mh *MsgHandler) Handle(req common.Msg) ([]interface{}, error) {
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

func (mh *MsgHandler) HandleBlocking(req common.Msg) (interface{}, *status.Status) {
	select {
	case mh.w.Mailbox() <- req:
	default:
		return nil, core.ErrInfo(codes.ResourceExhausted, "Worker busy",
			"capacity", cap(mh.w.Mailbox()))
	}
	select {
	case e := <-req.ResponseChannel():
		switch v := e.(type) {
		case *common.Response:
			r := e.(*common.Response)
			return r.Res, status.Convert(r.Err)
		default:
			return v, nil
		}
	case _ = <-time.After(time.Second * 5):
		return nil, core.ErrInfo(codes.DeadlineExceeded, "timeout",
			"msg", req)
	}
}
