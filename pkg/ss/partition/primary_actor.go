package partition

import (
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
	"reflect"
)

type primaryActor struct {
	actorBase
}

func (pa *primaryActor) become() error {
	pa.setState(Primary)
	pa.logger.Info("became", zap.Stringer("state", pa.getState()), zap.Int64("id", pa.id))
	followers := make([]common.Mailbox, 0)
channelRead:
	for m := range pa.mailBox {
		switch m.(type) {
		case *FollowRequest:
			fr := m.(*FollowRequest)
			followers = append(followers, fr.FollowerMailbox)
			pa.logger.Info("adding follower", zap.Int64("to part", pa.id), zap.Int64("follower id", fr.FollowerId))
			select {
			case fr.FollowerMailbox <- &common.AppStateImpl{S: pa.consumer.GetSnapshot()}:
			default:
				// todo
			}
		case common.ClientMsg:
			cm := m.(common.ClientMsg)
			handleClientMsg(&pa.actorBase, cm)
			if !cm.ReadOnly() && len(followers) > 0 {
				for _, f := range followers {
					select {
					case f <- &common.ReplicatedMsg{ClientMsg: cm}:
					default:
						// todo
					}
				}
			}
		case *common.ReplicatedMsg:
			handleReplicatedMsg(&pa.actorBase, m.(*common.ReplicatedMsg))
		case BecomeMsg:
			bm := m.(BecomeMsg)
			switch bm.Target() {
			case Proxy:
				bp := m.(*BecomeProxy)
				prx := &proxyActor{pa.actorBase, bp.ProxyTo}
				pa.logger.Info("becoming proxy", zap.Int64("part", pa.id))
				prx.setState(Proxy)
				return prx.become()
			case Retired:
				pa.logger.Info("retiring", zap.Int64("part", pa.id))
				break channelRead
			default:
			}
		default:
			pa.logger.Warn("not handled", zap.Int64("part", pa.id), zap.Stringer("state", pa.getState()), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	pa.setState(Retired)
	close(pa.mailBox)
	return nil
}
