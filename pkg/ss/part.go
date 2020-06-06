package ss

import dstk "github.com/anujga/dstk/pkg/api/proto"

type PartRange struct {
	partition *dstk.Partition
	consumer  PartHandler
	mailBox   chan Msg
}

func (p *PartRange) Start() []byte {
	return p.partition.GetStart()
}

func (p *PartRange) End() []byte {
	return p.partition.GetEnd()
}

func (p *PartRange) Run() bool {
	for m := range p.mailBox {
		p.consumer.Process(m)
	}
	return true
}
