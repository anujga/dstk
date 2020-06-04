package se

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
)

type slicerCli struct {
	cli      ThickClient
	connPool core.ConnPool
}

type PartId int64
type WorkerId int64

// Used by applications who want to implement their own thick clients
type ThickClient interface {

	// used at runtime for key lookup
	Get(ctx context.Context, key []byte) (*pb.Partition, error)

	// whenever cluster config changes, notifications are sent.
	// primary usecase would be to create / close connections
	// Based on usecases, will change the payload carried in
	// channel, for now we send Time
	Notifications() <-chan interface{}

	// as a response to notification, the client is likely to request
	// a snapshot. this is an expensive op, hence provided separately
	// instead of sending on the channel.
	Parts() []*pb.Partition
}
