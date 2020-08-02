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
	pids := make([]int64, 0)
	for _, p := range pa.proxyTo {
		pids = append(pids, p.Id())
	}
	pa.logger.Info("became", zap.String("state", pa.getState().String()), zap.Int64("id", pa.id), zap.Int64s("proxy to", pids))
	channelRead:
	for m := range pa.mailBox {
		switch m.(type) {
		case common.ClientMsg:
			cm := m.(common.ClientMsg)
			for _, a := range pa.proxyTo {
				if a.Contains(cm.Key()) {
					select {
					case a.Mailbox() <- &common.ProxiedMsg{ClientMsg: cm}:
					default:
						cm.ResponseChannel() <- core.ErrInfo(codes.ResourceExhausted, "Worker busy",
							"capacity", cap(a.Mailbox()))
					}
				}
			}
		case *Retire:
			pa.logger.Info("retiring", zap.Int64("part", pa.id))
			break channelRead
		default:
			pa.logger.Warn("not handled", zap.Int64("part", pa.id), zap.Any("state", pa.getState().String()), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	pa.setState(Retired)
	close(pa.mailBox)
	return nil
}
