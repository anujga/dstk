package partition

import (
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
	"reflect"
)

type catchingUpActor struct {
	actorBase
}

func (fa *catchingUpActor) become(leaderMailbox common.Mailbox) error {
	fa.setState(CatchingUp)
	fa.logger.Info("became", zap.String("smstate", fa.getState().String()), zap.Int64("id", fa.id))
	select {
	case leaderMailbox <- &FollowRequest{FollowerMailbox: fa.mailBox}:
	default:
		// todo
	}
	// todo pass capacity as a parameter
	msgList := make([]common.ClientMsg, 0)
	for m := range fa.mailBox {
		switch m.(type) {
		case common.AppState:
			if err := fa.consumer.ApplySnapshot(m.(common.AppState)); err == nil {
				for _, msg := range msgList {
					res, err := fa.consumer.Process(msg)
					resC := msg.ResponseChannel()
					if err != nil {
						resC <- err
					} else {
						resC <- res
					}
					close(resC)
				}
				fa := followingActor{fa.actorBase}
				return fa.become()
			} else {
				return err
			}
		case common.ClientMsg:
			if len(msgList) == 1024 {
				// todo handle
			}
			msgList = append(msgList, m.(common.ClientMsg))
		default:
			fa.logger.Warn("not handled", zap.Any("state", fa.smState), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	fa.setState(Completed)
	return nil
}
