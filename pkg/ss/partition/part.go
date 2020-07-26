package partition

import (
	"bytes"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
	"sync/atomic"
)

type Actor interface {
	Start() core.KeyT
	End() core.KeyT
	Mailbox() common.Mailbox
	Id() int64
	Stop()
	Run() *core.FutureErr
	CanServe() bool
	State() State
	Contains(k core.KeyT) bool
}

type actorImpl struct {
	partition *pb.Partition
	smState   *atomic.Value
	consumer  common.Consumer
	mailBox   chan interface{}
	Done      *core.FutureErr
	logger    *zap.Logger
}

func (p *actorImpl) Contains(k core.KeyT) bool {
	return bytes.Compare(p.Start(), k) <= 0 && bytes.Compare(k, p.End()) < 0
}

func (p *actorImpl) CanServe() bool {
	s := p.State()
	return s == Primary || s == Proxy
}

func (p *actorImpl) Mailbox() common.Mailbox {
	return p.mailBox
}

func (p *actorImpl) Start() core.KeyT {
	return p.partition.GetStart()
}

func (p *actorImpl) End() core.KeyT {
	return p.partition.GetEnd()
}

func (p *actorImpl) State() State {
	return p.smState.Load().(State)
}

func (p *actorImpl) Id() int64 {
	return p.partition.GetId()
}

func (p *actorImpl) Run() *core.FutureErr {
	// ensure state is not mutated in other threads
	ia := initActor{actorBase{
		id:      p.Id(),
		logger:  p.logger,
		smState: p.smState,
		mailBox: p.mailBox,
		consumer: p.consumer,
	}}
	return p.Done.Complete(ia.become)
}

//this will not be effective till the consumer
//has read all the messages from the channel
func (p *actorImpl) Stop() {
	close(p.mailBox)
}

func NewActor(p *pb.Partition, c common.Consumer, maxOutstanding int) Actor {
	ai := &actorImpl{
		partition: p,
		consumer:  c,
		smState:   &atomic.Value{},
		mailBox:   make(chan interface{}, maxOutstanding),
		Done:      core.NewPromise(),
		logger:    zap.L(),
	}
	ai.smState.Store(Init)
	return ai
}
