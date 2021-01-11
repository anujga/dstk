package partition

import (
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
	"reflect"
)

type catchingUpActor struct {
	actorBase
	leaderMailbox common.Mailbox
	leaderId      int64
}

func (fa *catchingUpActor) become() error {
	fa.logger.Info("became",
		zap.Stringer("state", fa.getState()),
		zap.Int64("id", fa.id),
		zap.Int64("leader id", fa.leaderId))
	select {
	case fa.leaderMailbox <- &FollowRequest{FollowerMailbox: fa.mailBox, FollowerId: fa.id}:
	default:
		// todo
	}
	fa.setState(CatchingUp)
	// todo pass capacity as a parameter
	msgList := make([]*common.ReplicatedMsg, 0)
	for m := range fa.mailBox {
		switch m.(type) {
		case common.AppState:
			if err := fa.consumer.ApplySnapshot(m.(common.AppState)); err == nil {
				for _, _ = range msgList {
					// todo no-op as leader is in same node
				}
				fa := followingActor{fa.actorBase}
				return fa.become()
			} else {
				return err
			}
		case *common.ReplicatedMsg:
			if len(msgList) == 1024 {
				// todo handle
			}
			msgList = append(msgList, m.(*common.ReplicatedMsg))
		default:
			fa.logger.Warn("not handled", zap.Stringer("state", fa.getState()), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	fa.setState(Retired)
	return nil
}
