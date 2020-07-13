package ss

import "context"

type FollowRequest struct {
	followerMailbox chan<- interface{}
}

type CtrlMsg struct {
	grpcReq interface{}
	ctx context.Context
	ch chan interface{}
}

func (c *CtrlMsg) ResponseChannel() chan interface{} {
	return c.ch
}

type FollowerCaughtup struct {
	p *PartRange
}