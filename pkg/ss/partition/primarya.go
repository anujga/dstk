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
	pa.logger.Info("became", zap.String("state", pa.getState().String()), zap.Int64("id", pa.id))
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
			prx := &proxyActor{pa.actorBase, bp.ProxyTo}
			pa.logger.Info("becoming proxy", zap.Int64("part", pa.id))
			prx.setState(Proxy)
			return prx.become()
		case *Retire:
			pa.logger.Info("retiring", zap.Int64("part", pa.id))
			break
		default:
			pa.logger.Warn("not handled", zap.Int64("part", pa.id), zap.Any("state", pa.getState().String()), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	pa.setState(Completed)
	return nil
}
