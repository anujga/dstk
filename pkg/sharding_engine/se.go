package se

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
)

type SliceRdr interface {
	Get(key []byte) (Partition, error)
}

type slicerCli struct {
	cli      SliceRdr
	connPool core.ConnPool
}

type PartId int64
type WorkerId int64

type Client interface {
	Lookup(ctx context.Context, id PartId) (pb.Partition, error)
}

//type ShardingEngine interface {
//	ClientApi
//	WorkerApi
//	AssignerApi
//}
