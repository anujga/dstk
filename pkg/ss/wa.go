package ss

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

type WorkerActor interface {
	Mailbox() chan<- interface{}
	Start() *core.FutureErr
}

type WaImpl struct {
	mailbox chan interface{}
	pm      *PartitionMgr
	logger *zap.Logger
}

func (w *WaImpl) Mailbox() chan<- interface{} {
	return w.mailbox
}

func (w *WaImpl) clientReq(msg ClientMsg) {
	p, err := w.pm.Find(msg.Key())
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
			case ClientMsg:
				w.clientReq(msg.(ClientMsg))
			case *CtrlMsg:
				w.ctrlReq(msg.(*CtrlMsg))
			case *SplitsCaughtup:
				w.splitsCaughtup(msg.(*SplitsCaughtup))
			default:
				w.logger.Warn("no handler", zap.Any("msg", msg))
			}
		}
		return nil
	})
	return fut
}

func (w *WaImpl) startNewPart(partition *pb.Partition, cl func(*PartRange)) (PartitionActor, error) {
	c, maxOutstanding, err := w.pm.consumer.Make(partition)
	if err != nil {
		return nil, err
	}
	return NewPartRange(partition, c, maxOutstanding, cl), nil
}

func (w *WaImpl) ctrlReq(msg *CtrlMsg) {
	switch msg.grpcReq.(type) {
	case *pb.SplitPartReq:
		sr := msg.grpcReq.(*pb.SplitPartReq)
		w.logger.Info("Handling split request",
			zap.Any("source", sr.GetSourcePartition()),
			zap.Any("splits", sr.GetTargetPartitions()))
		// todo handle error
		pr, err := w.pm.Find(sr.SourcePartition.GetStart())
		if err == nil {
			cl := w.splitHandler(sr.GetTargetPartitions().GetParts(), pr.(*PartRange), sr.GetId())
			followers := make([]PartitionActor, len(sr.GetTargetPartitions().GetParts()))
			fIdx := 0
			for _, p := range sr.TargetPartitions.GetParts() {
				// todo handle error
				part, _ := w.startNewPart(p, cl)
				part.Run()
				followers[fIdx] = part
				fIdx++
			}
			select {
			case pr.Mailbox() <- &FollowRequest{followers: followers}:
			default:
				// todo handle
			}

		} else {
			msg.ResponseChannel() <- err
		}
		close(msg.ResponseChannel())
	}
}

func (w *WaImpl) splitHandler(tgtParts []*pb.Partition, parentRange *PartRange, id int64) func(partRange *PartRange) {
	m := make(map[int64]bool)
	splits := make([]*PartRange, 0)
	for _, tp := range tgtParts {
		m[tp.GetId()] = true
	}
	lk := &sync.Mutex{}
	l := w.logger.With(zap.Int64("split req id", id))
	return func(partRange *PartRange) {
		w.logger.Info("Caught up", zap.Int64("partid", partRange.partition.GetId()))
		lk.Lock()
		{
			delete(m, partRange.partition.GetId())
			splits = append(splits, partRange)
			if len(m) == 0 {
				l.Info("all splits caught up", zap.Any("splits", splits))
				select {
					case w.Mailbox() <- &SplitsCaughtup{
						splitRanges: splits,
						parentRange: parentRange,
					}:
				default:
					l.Error("failed to write splits caught up request on worker")
					// todo handle
				}
			}
		}
		lk.Unlock()
	}
}

func (w *WaImpl) splitsCaughtup(caughtup *SplitsCaughtup) error {
	pmState := w.pm.State()
	// todo handle error
	var err error
	_, err = pmState.removePart(caughtup.parentRange)
	if err != nil {
		return err
	}
	for _, pr := range caughtup.splitRanges {
		// todo handler error. should we revert in that case?
		if err = pmState.addPart(pr); err != nil {
			break
		}
	}
	if err != nil {
		// todo revert
	}
	w.logger.Info("splits are updated in partition manager state")
	// todo update this status
	return err
}

//todo: ensure there is at least 1 partition during construction
func NewPartitionMgr2(workerId se.WorkerId, consumer ConsumerFactory, rpc pb.SeWorkerApiClient, maker func() interface{}) (WorkerActor, *status.Status) {
	pm := &PartitionMgr{
		consumer:       consumer,
		rpc:            rpc,
		id:             workerId,
		slog:           zap.S().With("workerId", workerId),
		initStateMaker: maker,
	}

	rep := core.Repeat(5*time.Hour, func(timestamp time.Time) bool {
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
	}, true)

	if rep == nil {
		return nil, core.ErrInfo(
			codes.Internal,
			"failed to initialize via se",
			"se", rpc)
	}
	return &WaImpl{
		// take size as param
		mailbox: make(chan interface{}, 10000),
		pm:      pm,
		logger: zap.L(),
	}, nil
}
