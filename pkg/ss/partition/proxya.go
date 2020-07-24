package partition

import (
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"reflect"
)

type proxyActor struct {
	actorBase
}

func (pa *proxyActor) become(mb common.Mailbox) error {
	pa.setState(Proxy)
	pa.logger.Info("became", zap.String("smstate", pa.getState().String()), zap.Int64("id", pa.id))
	for m := range pa.mailBox {
		switch m.(type) {
		case common.ClientMsg:
			select {
			case mb <- m:
			default:
				cm := m.(common.ClientMsg)
				cm.ResponseChannel() <- core.ErrInfo(codes.ResourceExhausted, "Worker busy",
					"capacity", cap(mb))
			}
		default:
			pa.logger.Warn("not handled", zap.Any("state", pa.smState), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	pa.setState(Completed)
	return nil
}
