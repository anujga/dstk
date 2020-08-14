package partition

import (
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
	"reflect"
)

type followingActor struct {
	actorBase
}

func (fa *followingActor) become() error {
	fa.setState(Follower)
	fa.logger.Info("became", zap.Stringer("state", fa.getState()), zap.Int64("id", fa.id))
	for m := range fa.mailBox {
		switch m.(type) {
		case BecomeMsg:
			bm := m.(BecomeMsg)
			switch bm.Target() {
			case Primary:
				pa := &primaryActor{fa.actorBase}
				return pa.become()
			default:
			}
		case *common.ProxiedMsg:
			// todo it looks a bit odd for follower to process client messages, but this is ensuring
			// the correctness of algorithm. we can revisit this.
			pm := m.(*common.ProxiedMsg)
			handleClientMsg(&fa.actorBase, pm.ClientMsg)
		case *common.ReplicatedMsg:
			handleReplicatedMsg(&fa.actorBase, m.(*common.ReplicatedMsg))
		default:
			fa.logger.Warn("not handled", zap.Int64("part", fa.id), zap.Stringer("state", fa.getState()), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	fa.setState(Retired)
	return nil
}
