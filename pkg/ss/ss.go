package ss

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
)

type Msg interface {
	ResponseChannel() chan interface{}
}

type ClientMsg interface {
	Msg
	ReadOnly() bool
	Key() core.KeyT
}

type AppState interface {
	State() interface{}
}

type PartHandler interface {
	Process(msg Msg) (interface{}, error)
	GetSnapshot() AppState
	ApplySnapshot(as AppState) error
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
