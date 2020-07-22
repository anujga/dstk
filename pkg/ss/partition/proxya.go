package partition

import (
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"reflect"
)

type proxyActor struct {
	*actorImpl
}

func (pa *proxyActor) become(primaryActor primaryActor) error {
	pa.smState = Proxy
	pa.logger.Info("became", zap.String("smstate", pa.smState.String()), zap.Int64("id", pa.Id()))
	for m := range pa.mailBox {
		switch m.(type) {
		case common.ClientMsg:
			select {
			case primaryActor.Mailbox() <- m:
			default:
				cm := m.(common.ClientMsg)
				cm.ResponseChannel() <- core.ErrInfo(codes.ResourceExhausted, "Worker busy",
					"capacity", cap(primaryActor.Mailbox()))
			}

		default:
			pa.logger.Warn("not handled", zap.Any("state", pa.smState), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	pa.smState = Completed
	return nil
}
