package ss

type FollowRequest struct {
	followerMailbox chan<- interface{}
}

