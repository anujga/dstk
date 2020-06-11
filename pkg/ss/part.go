package ss

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
)

type PartRange struct {
	partition *pb.Partition
	consumer  PartHandler
	mailBox   chan Msg
	Done      *core.FutureErr
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
	return p.Done.Complete(func() error {
		for m := range p.mailBox {
			p.consumer.Process(m)
		}
		return nil
	})
}

//this will not be effective till the consumer
//has read all the messages from the channel
func (p *PartRange) Stop() {
	close(p.mailBox)
}

func NewPartRange(p *pb.Partition, c PartHandler, maxOutstanding int) *PartRange {
	return &PartRange{
		partition: p,
		consumer:  c,
		mailBox:   make(chan Msg, maxOutstanding),
		Done:      core.NewPromise(),
	}
}
