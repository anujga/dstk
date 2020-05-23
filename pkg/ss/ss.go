package ss

import dstk "github.com/anujga/dstk/pkg/api/proto"

type KeyT []byte

type Msg interface {
	ReadOnly() bool
	Key() KeyT
}

type Consumer interface {
	Process(msg Msg) bool
	Meta() *dstk.Partition
	MaxOutstanding() int
}
