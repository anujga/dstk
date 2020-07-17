package partrpc

import (
	"context"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"google.golang.org/grpc"
)

type PartitionClientPool interface {
	GetClient(ctx context.Context, key []byte) (PartitionClient, error)
}

type clientPool struct {
	tc   se.ThickClient
	pool core.ConnPool
}

func (c *clientPool) GetClient(ctx context.Context, key []byte) (PartitionClient, error) {
	part, err := c.tc.Get(ctx, key)

	if err != nil {
		return nil, err
	}

	pc, err := c.pool.Get(ctx, part.GetUrl())
	if err != nil {
		return nil, err
	}

	return pc.(PartitionClient), nil
}

func NewPartitionClientPool(rpcClientFactory RpcClientFactory, seClient dstk.SeClientApiClient, connectionOpts ...grpc.DialOption) PartitionClientPool {
	tc := se.NewThickClient("c1", seClient)
	// wait till state syncs once. any other better way?
	//time.Sleep(time.Second*80)
	return &clientPool{pool: core.NonExpiryPool(newRpcConnFactory(rpcClientFactory, connectionOpts...)), tc: tc}
}
