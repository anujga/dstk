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
		case *BecomeCatchingUpActor:
			fm := m.(*BecomeCatchingUpActor)
			ca := &catchingUpActor{ia.actorBase, fm.LeaderMailbox}
			return ca.become()
		default:
			ia.logger.Warn("not handled", zap.Any("state", ia.getState().String()), zap.Any("type", reflect.TypeOf(m)), zap.Int64("part", ia.id))
		}
	}
	return nil
}
