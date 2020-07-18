package ss

import "context"

type FollowRequest struct {
	followers []PartitionActor
}

type CtrlMsg struct {
	grpcReq interface{}
	ctx context.Context
	ch chan interface{}
}

func (c *CtrlMsg) ResponseChannel() chan interface{} {
	return c.ch
}

type SplitsCaughtup struct {
	parentRange *PartRange
	splitRanges []*PartRange
}

type PurgePart struct {
	pa PartitionActor
}