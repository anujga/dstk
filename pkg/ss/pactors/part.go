package pactors

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
)

type State int


type PartitionActor interface {
	Start() core.KeyT
	End() core.KeyT
	Mailbox() chan<- interface{}
	Id() int64
	Stop()
	Run() *core.FutureErr
	CanServe() bool
}

type PartRange struct {
	partition *pb.Partition
	smState   State
	consumer  common.Consumer
	mailBox   chan interface{}
	Done      *core.FutureErr
	logger    *zap.Logger
	leader    PartitionActor
}

func (p *PartRange) CanServe() bool {
	return p.smState == Primary || p.smState == Proxy
}

func (p *PartRange) Mailbox() chan<- interface{} {
	return p.mailBox
}

func (p *PartRange) Start() core.KeyT {
	return p.partition.GetStart()
}

func (p *PartRange) End() core.KeyT {
	return p.partition.GetEnd()
}

func (p *PartRange) Id() int64 {
	return p.partition.GetId()
}

func (p *PartRange) Run() *core.FutureErr {
	// ensure state is not mutated in other threads
	ia := initActor{p}
	return p.Done.Complete(ia.become)
}

//this will not be effective till the consumer
//has read all the messages from the channel
func (p *PartRange) Stop() {
	close(p.mailBox)
}

func NewPartActor(p *pb.Partition, c common.Consumer, maxOutstanding int, leader PartitionActor) PartitionActor {
	return &PartRange{
		partition: p,
		smState:   Init,
		consumer:  c,
		mailBox:   make(chan interface{}, maxOutstanding),
		Done:      core.NewPromise(),
		leader:    leader,
		logger:    zap.L(),
	}
}
