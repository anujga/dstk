package partition

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
)

type State int


type Actor interface {
	Start() core.KeyT
	End() core.KeyT
	Mailbox() chan<- interface{}
	Id() int64
	Stop()
	Run() *core.FutureErr
	CanServe() bool
}

type actorImpl struct {
	partition *pb.Partition
	smState   State
	consumer  common.Consumer
	mailBox   chan interface{}
	Done      *core.FutureErr
	logger    *zap.Logger
	leader    Actor
}

func (p *actorImpl) CanServe() bool {
	return p.smState == Primary || p.smState == Proxy
}

func (p *actorImpl) Mailbox() chan<- interface{} {
	return p.mailBox
}

func (p *actorImpl) Start() core.KeyT {
	return p.partition.GetStart()
}

func (p *actorImpl) End() core.KeyT {
	return p.partition.GetEnd()
}

func (p *actorImpl) Id() int64 {
	return p.partition.GetId()
}

func (p *actorImpl) Run() *core.FutureErr {
	// ensure state is not mutated in other threads
	ia := initActor{p}
	return p.Done.Complete(ia.become)
}

//this will not be effective till the consumer
//has read all the messages from the channel
func (p *actorImpl) Stop() {
	close(p.mailBox)
}

func NewActor(p *pb.Partition, c common.Consumer, maxOutstanding int, leader Actor) Actor {
	return &actorImpl{
		partition: p,
		smState:   Init,
		consumer:  c,
		mailBox:   make(chan interface{}, maxOutstanding),
		Done:      core.NewPromise(),
		leader:    leader,
		logger:    zap.L(),
	}
}
