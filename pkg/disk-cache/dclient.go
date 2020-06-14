package diskcache

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"github.com/anujga/dstk/pkg/ss/partrpc"
	"google.golang.org/grpc"
)

type Client interface {
	Get(key core.KeyT) ([]byte, error)
	Put(key, value core.KeyT, ttlSeconds float32) error
	Remove(key core.KeyT) error
}

type impl struct {
	clientPool partrpc.PartitionClientPool
}

func (i *impl) Get(key core.KeyT) ([]byte, error) {
	ctx := context.TODO()
	if client, err := i.clientPool.GetClient(ctx, key); err == nil {
		dcClient := client.RpcClient().(pb.DcRpcClient)
		if rpcResponse, err := dcClient.Get(ctx, &pb.DcGetReq{Key: key}); err == nil {
			return rpcResponse.GetValue(), nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (i *impl) Put(key, value core.KeyT, ttlSeconds float32) error {
	panic("implement me")
}

func (i *impl) Remove(key core.KeyT) error {
	panic("implement me")
}

func NewClient(ctx context.Context, seUrl string, opts ...grpc.DialOption) (Client, error) {
	seClient, err := se.NewSeClient(ctx, seUrl, opts...)
	if err != nil {
		return nil, err
	}
	clientPool := partrpc.NewPartitionClientPool(func(conn *grpc.ClientConn) interface{} {
		return pb.NewDcRpcClient(conn)
	}, seClient, opts...)
	return &impl{
		clientPool: clientPool,
	}, nil
}
