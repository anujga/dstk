package common

import (
	"github.com/anujga/dstk/pkg/core"
)

type Mailbox chan<- interface{}

type Msg interface {
	ResponseChannel() chan interface{}
}

type ClientMsg interface {
	Msg
	ReadOnly() bool
	Key() core.KeyT
}

type Response struct {
	Res interface{}
	Err error
}

// todo ideally, we don't want to differentiate between replicated message and a client message.
// but since the messages are already applied by the primary actor, we want to differentiate the two
// so that the actor receiving this can treat it as a no-op
type ReplicatedMsg struct {
	ClientMsg
}

type ProxiedMsg struct {
	ClientMsg
}
