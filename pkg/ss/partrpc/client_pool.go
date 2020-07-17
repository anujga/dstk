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
		return nil, err
	}

	pc, err := c.pool.Get(ctx, part.GetUrl())
	if err != nil {
		return nil, err
	}

	return pc.(PartitionClient), nil
}

func NewPartitionClientPool(clientId string, rpcClientFactory RpcClientFactory, seClient dstk.SeClientApiClient, connectionOpts ...grpc.DialOption) (PartitionClientPool, *status.Status) {
	tc, err := se.NewThickClient(clientId, seClient)
	if err != nil {
		return nil, err
	}
	factory := newRpcConnFactory(rpcClientFactory, connectionOpts...)
	pool := core.NonExpiryPool(factory)
	return &clientPool{pool: pool, tc: tc}, nil
}
