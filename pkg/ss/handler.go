package ss

import (
	"errors"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/core/control"
	"github.com/anujga/dstk/pkg/ss/common"
	"github.com/anujga/dstk/pkg/ss/node"
	"google.golang.org/grpc/codes"
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

func (mh *MsgHandler) HandleBlocking(req common.Msg) *control.Response {
	select {
	case mh.w.Mailbox() <- req:
	default:
		err := core.ErrInfo(codes.ResourceExhausted, "Worker busy",
			"capacity", cap(mh.w.Mailbox()))
		return control.Failure(err)
	}
	select {
	case e := <-req.ResponseChannel():
		return e
	case _ = <-time.After(time.Second * 5):
		err := core.ErrInfo(codes.DeadlineExceeded, "timeout",
			"msg", req)
		return control.Failure(err)
	}
}
