package common

import (
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
