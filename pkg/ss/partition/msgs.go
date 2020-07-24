package partition

import "github.com/anujga/dstk/pkg/ss/common"

type FollowRequest struct {
	FollowerId      int64
	FollowerMailbox common.Mailbox
}

type BecomePrimary struct {
}

type BecomeFollower struct {
	LeaderId int64
	LeaderMailbox common.Mailbox
}

type BecomeProxy struct {
	ProxyToId int64
	ProxyTo   common.Mailbox
}

type Retire struct {
}
