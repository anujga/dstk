package partition

import "github.com/anujga/dstk/pkg/ss/common"

type FollowRequest struct {
	FollowerId      int64
	FollowerMailbox common.Mailbox
}

type BecomeMsg interface {
	Target() State
}

type BecomeMsgImpl struct {
	TargetState State
}

func (b *BecomeMsgImpl) Target() State {
	return b.TargetState
}

type BecomeCatchingUpActor struct {
	LeaderId      int64
	LeaderMailbox common.Mailbox
}

func (b *BecomeCatchingUpActor) Target() State {
	return CatchingUp
}

type BecomeProxy struct {
	ProxyTo []Actor
}

func (b *BecomeProxy) Target() State {
	return Proxy
}
