package partition

import (
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
	"reflect"
)

type catchingUpActor struct {
	*PartRange
}

func (fa *catchingUpActor) become() error {
	fa.smState = CatchingUp
	fa.logger.Info("became", zap.String("smstate", fa.smState.String()), zap.Int64("id", fa.Id()))
	fa.leader.Mailbox() <- &FollowRequest{Follower: fa}
	// todo pass capacity as a parameter
	msgList := make([]common.ClientMsg, 0, 1024)
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
				fa := followingActor{fa.PartRange}
				return fa.become()
			} else {
				return err
			}
		case common.ClientMsg:
			if len(msgList) == cap(msgList) {
				// todo handle
			}
			msgList[len(msgList)] = m.(common.ClientMsg)
		default:
			fa.logger.Warn("not handled", zap.Any("state", fa.smState), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	fa.smState = Completed
	return nil
}
