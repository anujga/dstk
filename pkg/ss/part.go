package ss

import (
	"errors"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
)

type State int

const (
	Init State = iota
	Loading
	Running
	Completed
)

type PartitionActor interface {
	Mailbox() chan<- Msg
	Id() int64
}

type PartRange struct {
	smState   State
	partition *pb.Partition
	consumer  PartHandler
	mailBox   chan Msg
	Done      *core.FutureErr
}

func (p *PartRange) Mailbox() chan<- Msg {
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

func (p *PartRange) becomeRunningHandler() error {
	p.smState = Running
	for m := range p.mailBox {
		p.consumer.Process(m)
	}
	return nil
}

func (p *PartRange) becomeLoadingHandler() error {
	p.smState = Loading
	s := <-p.mailBox
	var err error
	if appState, ok := s.(AppState); ok {
		if err = p.consumer.ApplySnapshot(appState); err == nil {
			return p.becomeRunningHandler()
		}
	} else {
		err = errors.New("invalid message received")
	}
	return err
}

func (p *PartRange) Run() *core.FutureErr {
	// ensure state is not mutated in other threads
	return p.Done.Complete(p.becomeLoadingHandler)
}

//this will not be effective till the consumer
//has read all the messages from the channel
func (p *PartRange) Stop() {
	close(p.mailBox)
	p.smState = Completed
}

func NewPartRange(p *pb.Partition, c PartHandler, maxOutstanding int) *PartRange {
	return &PartRange{
		smState:   Init,
		partition: p,
		consumer:  c,
		mailBox:   make(chan Msg, maxOutstanding),
		Done:      core.NewPromise(),
	}
}
