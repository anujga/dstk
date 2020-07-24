package partition

import (
	"go.uber.org/zap"
	"reflect"
)

type initActor struct {
	actorBase
}

func (ia *initActor) become() error {
	for m := range ia.mailBox {
		switch m.(type) {
		case *BecomePrimary:
			pa := &primaryActor{ia.actorBase}
			return pa.become()
		case *BecomeFollower:
			fm := m.(*BecomeFollower)
			ca := &catchingUpActor{ia.actorBase}
			return ca.become(fm.LeaderMailbox)
		default:
			ia.logger.Warn("not handled", zap.Any("state", ia.smState), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	return nil
}
