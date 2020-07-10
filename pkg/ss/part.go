package ss

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"go.uber.org/zap"
	"reflect"
)

type State int

const (
	Init State = iota
	Loading
	Running
	Completed
)

type PartitionActor interface {
	Mailbox() chan<- interface{}
	Id() int64
}

type PartRange struct {
	smState   State
	partition *pb.Partition
	consumer  PartHandler
	mailBox   chan interface{}
	Done      *core.FutureErr
	logger    zap.Logger
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

func (p *PartRange) becomeRunningHandler() error {
	p.smState = Running
	var followerMailbox chan<- interface{}
	for m := range p.mailBox {
		switch m.(type) {
		case ClientMsg:
			cm := m.(ClientMsg)
			p.consumer.Process(cm)
			if !m.(ClientMsg).ReadOnly() && followerMailbox != nil {
				followerMailbox <- cm
			}
		case *FollowRequest:
			fr := m.(*FollowRequest)
			followerMailbox = fr.followerMailbox
			followerMailbox <- p.consumer.GetSnapshot()
		default:
			p.logger.Info("not handled", zap.Any("state", p.smState), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	return nil
}

func (p *PartRange) becomeLoadingHandler() error {
	p.smState = Loading
	// pass capacity as a parameter
	msgList := make([]ClientMsg, 1024)
	for m := range p.mailBox {
		switch m.(type) {
		case AppState:
			if err := p.consumer.ApplySnapshot(m.(AppState)); err == nil {
				for _, msg := range msgList {
					p.consumer.Process(msg)
				}
				return p.becomeRunningHandler()
			} else {
				return err
			}
		case ClientMsg:
			if len(msgList) == cap(msgList) {
				// todo handle
			}
			msgList[len(msgList)] = m.(ClientMsg)
		default:
			p.logger.Info("not handled", zap.Any("state", p.smState), zap.Any("type", reflect.TypeOf(m)))
		}
	}
	return nil
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
		mailBox:   make(chan interface{}, maxOutstanding),
		Done:      core.NewPromise(),
	}
}
