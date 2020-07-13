package ss

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/rangemap"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"time"
)

type WorkerActor interface {
	Mailbox() chan<- interface{}
	Start() *core.FutureErr
}

type WaImpl struct {
	mailbox chan interface{}
	pm      *PartitionMgr
}

func (w *WaImpl) Mailbox() chan<- interface{} {
	return w.mailbox
}

func (w *WaImpl) clientReq(msg ClientMsg) {
	p, err := w.pm.Find(msg.Key())
	if err != nil {
		msg.ResponseChannel() <- err
		close(msg.ResponseChannel())
	}
	select {
	case p.Mailbox() <- msg:
	default:
		msg.ResponseChannel() <- core.ErrInfo(codes.ResourceExhausted, "Partition Busy",
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
			case ClientMsg:
				w.clientReq(msg.(ClientMsg))
			case *CtrlMsg:
				w.ctrlReq(msg.(*CtrlMsg))
			case *FollowerCaughtup:
				w.caughtUp(msg.(*FollowerCaughtup))
			default:
				// todo handle this
			}
		}
		return nil
	})
	return fut
}

func (w *WaImpl) ctrlReq(msg *CtrlMsg) {
	switch msg.grpcReq.(type) {
	case *pb.SplitPartReq:
		// todo
	}
}

func (w *WaImpl) caughtUp(caughtup *FollowerCaughtup) {
	// todo
}

//todo: ensure there is at least 1 partition during construction
func NewPartitionMgr2(workerId se.WorkerId, consumer ConsumerFactory, rpc pb.SeWorkerApiClient, maker func() interface{}) WorkerActor {
	pm := &PartitionMgr{
		consumer:       consumer,
		rpc:            rpc,
		id:             workerId,
		slog:           zap.S().With("workerId", workerId),
		initStateMaker: maker,
	}

	core.Repeat(5*time.Second, func(timestamp time.Time) bool {
		err := pm.syncSe()
		if err != nil {
			pm.slog.Errorw("fetch updates from SE",
				"err", err)
		} else {
			delay := timestamp.UnixNano() - pm.State().lastModified
			pm.slog.Infow("fetch updates from SE",
				"time", timestamp,
				"delay", delay)
		}
		return true
	})
	pm.ResetMap(&state{
		m:            rangemap.New(15),
		lastModified: 0,
	})
	return &WaImpl{
		// take size as param
		mailbox: make(chan interface{}, 10000),
		pm:      pm,
	}
}
