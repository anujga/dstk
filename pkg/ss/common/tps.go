package common

import (
	"github.com/anujga/dstk/pkg/core"
)

type Mailbox chan interface{}

type Msg interface {
	ResponseChannel() Mailbox
}

type ClientMsg interface {
	Msg
	ReadOnly() bool
	Key() core.KeyT
}
