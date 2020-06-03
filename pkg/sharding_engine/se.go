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
	Get(ctx context.Context, key []byte) (*pb.Partition, error)
}
