package partition

import (
	"go.uber.org/zap"
	"reflect"
)

type initActor struct {
	actorBase
}

func (ia *initActor) become() error {
	ia.setState(Init)
	for m := range ia.mailBox {
		switch m.(type) {
		case BecomeMsg:
			bm := m.(BecomeMsg)
			switch bm.Target() {
			case Primary:
				pa := &primaryActor{ia.actorBase}
				return pa.become()
			default:
			}
		case *BecomeCatchingUpActor:
			fm := m.(*BecomeCatchingUpActor)
			ca := &catchingUpActor{ia.actorBase, fm.LeaderMailbox, fm.LeaderId}
			return ca.become()
		default:
			ia.logger.Warn("not handled", zap.Stringer("state", ia.getState()), zap.Any("type", reflect.TypeOf(m)), zap.Int64("part", ia.id))
		}
	}
	return nil
}
