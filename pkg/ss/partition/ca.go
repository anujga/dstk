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
	fa.logger.Info("became", zap.String("state", fa.getState().String()), zap.Int64("id", fa.id), zap.Int64("leader id", fa.leaderId))
	select {
	case fa.leaderMailbox <- &FollowRequest{FollowerMailbox: fa.mailBox, FollowerId: fa.id}:
	default:
		// todo
	}
	// todo pass capacity as a parameter
	msgList := make([]common.ClientMsg, 0)
	snapshotReceived := false
	for m := range fa.mailBox {
		switch m.(type) {
		case common.AppState:
			if err := fa.consumer.ApplySnapshot(m.(common.AppState)); err == nil {
				for _, _ = range msgList {
					// todo no-op as leader is in same node
				}
			} else {
				return err
			}
			snapshotReceived = true
		case common.ClientMsg:
			if snapshotReceived {
				// todo no-op as leader is in same node
			} else {
				if len(msgList) == 1024 {
					// todo handle
				}
				msgList = append(msgList, m.(common.ClientMsg))
			}
		case *BecomeFollower:
			if snapshotReceived {
				fa := followingActor{fa.actorBase}
				fa.setState(Follower)
				return fa.become()
			} else {
				fa.logger.Info("cannot become follower before receiving snapshot", zap.Int64("part", fa.id))
			}
		default:
			fa.logger.Warn("not handled", zap.Any("state", fa.getState().String()), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	fa.setState(Retired)
	return nil
}
