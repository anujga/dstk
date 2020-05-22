package ss

import dstk "github.com/anujga/dstk/pkg/api/proto"

type MsgTrait interface {
	ReadOnly() bool
}


type Consumer interface {
	Process(msg MsgTrait) bool
	Meta() dstk.Partition
}
