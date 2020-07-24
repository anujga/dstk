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
	pa.smState.Store(Primary)
	pa.logger.Info("became", zap.String("smstate", pa.getState().String()), zap.Int64("id", pa.id))
	followers := make([]common.Mailbox, 0)
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
		case *common.ClientMsg:
			cm := m.(common.ClientMsg)
			res, err := pa.consumer.Process(cm)
			resC := cm.ResponseChannel()
			if err != nil {
				resC <- err
			} else {
				resC <- res
			}
			close(resC)
			if !cm.ReadOnly() && len(followers) > 0 {
				for _, f := range followers {
					select {
					case f <- cm:
					default:
						// todo
					}
				}
			}
		case *BecomeProxy:
			bp := m.(*BecomeProxy)
			prx := &proxyActor{pa.actorBase}
			pa.logger.Info("becoming proxy", zap.Int64("part", pa.id))
			return prx.become(bp.ProxyTo)
		default:
			pa.logger.Warn("not handled", zap.Any("state", pa.smState), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	pa.setState(Completed)
	return nil
}
