package ss

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
)

type Msg interface {
	ReadOnly() bool
	Key() core.KeyT
	ResponseChannel() chan interface{}
}

type PartHandler interface {
	Process(msg Msg) bool
	//Meta() *dstk.Partition
}

type ConsumerFactory interface {
	Make(p *dstk.Partition) (PartHandler, int, error)
}

type Router interface {
	OnMsg(m Msg) error
}

type PartMgr interface {
	Find(key core.KeyT) *PartRange
	Add(p *dstk.Partition) error
}
