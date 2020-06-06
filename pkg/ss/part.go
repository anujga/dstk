package ss

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
)

type PartRange struct {
	partition *dstk.Partition
	consumer  PartHandler
	mailBox   chan Msg
}

func (p *PartRange) Start() core.KeyT {
	return p.partition.GetStart()
}

func (p *PartRange) End() core.KeyT {
	return p.partition.GetEnd()
}

func (p *PartRange) Run() bool {
	for m := range p.mailBox {
		p.consumer.Process(m)
	}
	return true
}
