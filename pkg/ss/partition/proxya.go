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
	proxyTo  []Actor
}

func (pa *proxyActor) become() error {
	pa.setState(Proxy)
	pa.logger.Info("became", zap.String("smstate", pa.getState().String()), zap.Int64("id", pa.id))
	for m := range pa.mailBox {
		switch m.(type) {
		case common.ClientMsg:
			cm := m.(common.ClientMsg)
			for _, a := range pa.proxyTo {
				if a.Contains(cm.Key()) {
					select {
					case a.Mailbox() <- cm:
					default:
						cm.ResponseChannel() <- core.ErrInfo(codes.ResourceExhausted, "Worker busy",
							"capacity", cap(a.Mailbox()))
					}
				}
			}
		default:
			pa.logger.Warn("not handled", zap.Int64("part", pa.id), zap.Any("state", pa.getState().String()), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	pa.setState(Completed)
	return nil
}
