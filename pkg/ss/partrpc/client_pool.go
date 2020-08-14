package partrpc

import (
	"context"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
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
		return nil, err.Err()
	}

	pc, er2 := c.pool.Get(ctx, part.GetUrl())
	if er2 != nil {
		return nil, er2
	}

	return pc.(PartitionClient), nil
}

func NewPartitionClientPool(clientId string, rpcClientFactory RpcClientFactory, seClient dstk.PartitionRpcClient, connectionOpts ...grpc.DialOption) (PartitionClientPool, *status.Status) {
	tc, err := se.NewThickClient(clientId, seClient)
	if err != nil {
		return nil, err
	}
	factory := newRpcConnFactory(rpcClientFactory, connectionOpts...)
	pool := core.NonExpiryPool(factory)
	return &clientPool{pool: pool, tc: tc}, nil
}
