package node

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"github.com/anujga/dstk/pkg/ss/common"
	"github.com/anujga/dstk/pkg/ss/partmgr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

type WorkerActor interface {
	Mailbox() chan<- interface{}
	Start() *core.FutureErr
	Id() se.WorkerId
}

type WaImpl struct {
	id      se.WorkerId
	mailbox chan interface{}
	partMgr partmgr.PartManager
	logger  *zap.Logger
}

func (w *WaImpl) Id() se.WorkerId {
	return w.id
}

func (w *WaImpl) Mailbox() chan<- interface{} {
	return w.mailbox
}

func (w *WaImpl) clientReq(msg common.ClientMsg) {
	p, err := w.partMgr.Find(msg.Key())
	if err != nil {
		msg.ResponseChannel() <- err
		close(msg.ResponseChannel())
		return
	}
	if p == nil {
		panic("partition should not be null here")
	}
	select {
	case p.Mailbox() <- msg:
	default:
		msg.ResponseChannel() <- core.ErrInfo(
			codes.ResourceExhausted, "Partition Busy",
			"capacity", cap(p.Mailbox()),
			"partition", p.Id()).Err()
		close(msg.ResponseChannel())
	}
}

// Single threaded router. 1 channel per partition
// path=data
func (w *WaImpl) Start() *core.FutureErr {
	fut := core.NewPromise()
	fut.Complete(func() error {
		// ensure state is not mutated in other threads
		for msg := range w.mailbox {
			switch msg.(type) {
			case common.ClientMsg:
				w.clientReq(msg.(common.ClientMsg))
			case *pb.PartList:
				// todo handle error
				_ = w.partMgr.Reset(msg.(*pb.PartList))
			default:
				w.logger.Warn("no handler", zap.Any("msg", msg))
			}
		}
		return nil
	})
	return fut
}

func NewWorker(factory common.ConsumerFactory, id se.WorkerId) (WorkerActor, error) {
	p, err := partmgr.NewPartitionMgr(factory)
	if err != nil {
		return nil, err.Err()
	}
	w := &WaImpl{
		// take size as param
		mailbox: make(chan interface{}, 10000),
		partMgr: p,
		logger:  zap.L(),
		id:      id,
	}
	return w, nil
}
