package ss

import dstk "github.com/anujga/dstk/pkg/api/proto"

type KeyT []byte

type Msg interface {
	ReadOnly() bool
	Key() KeyT
	ResponseChannel() chan interface{}
}

type Consumer interface {
	Process(ptask Msg) bool
	//Meta() *dstk.Partition
}

type ConsumerFactory interface {
	Make(p *dstk.Partition) (Consumer, int, error)
}

type Router interface {
	OnMsg(m Msg) error
}

type PartMgr interface {
	Find(key KeyT) *PartItem
	Add(p *dstk.Partition) error
}
