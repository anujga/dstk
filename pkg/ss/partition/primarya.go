package partition

import (
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
	"reflect"
)

type primaryActor struct {
	*actorImpl
}

func (pa *primaryActor) become() error {
	pa.smState = Primary
	pa.logger.Info("became", zap.String("smstate", pa.smState.String()), zap.Int64("id", pa.Id()))
	followers := make([]Actor, 0)
	for m := range pa.mailBox {
		switch m.(type) {
		case *FollowRequest:
			fr := m.(*FollowRequest)
			followers = append(followers, fr.Follower)
			pa.logger.Info("adding follower", zap.Int64("added part", fr.Follower.Id()), zap.Int64("to part", pa.Id()))
			fr.Follower.Mailbox() <- &common.AppStateImpl{S: pa.consumer.GetSnapshot()}
		default:
			pa.logger.Warn("not handled", zap.Any("state", pa.smState), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	pa.smState = Completed
	return nil
}