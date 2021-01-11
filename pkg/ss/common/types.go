package common

import (
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/core/control"
)

type Mailbox chan<- interface{}

type Msg interface {
	//todo: writing to this msg can fail. wrap that up
	ResponseChannel() chan *control.Response
}

type ClientMsg interface {
	Msg
	ReadOnly() bool
	Key() core.KeyT
}

// todo ideally, we don't want to differentiate between replicated message and a client message.
// but since the messages are already applied by the primary actor, we want to differentiate the two
// so that the actor receiving this can treat it as a no-op
type ReplicatedMsg struct {
	ClientMsg
}

// todo add an enum to differentiate between client msg, replicated, proxied message
type ProxiedMsg struct {
	ClientMsg
}
