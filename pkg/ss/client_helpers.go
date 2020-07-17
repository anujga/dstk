package ss

import (
	"context"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/partrpc"
	"google.golang.org/grpc"
)

type ClientBase struct {
	Pool partrpc.PartitionClientPool
}

func (i *ClientBase) Fwd(ctx context.Context, key core.KeyT, in interface{}, out interface{}, api string, opts ...grpc.CallOption) error {

	rpc, err := i.RawClient(ctx, key)
	if err != nil {
		return err
	}

	err = rpc.Invoke(ctx, api, in, out, opts...)
	if err != nil {
		return err
	}
	return nil
}

func (i *ClientBase) RawClient(ctx context.Context, k []byte) (grpc.ClientConnInterface, error) {
	client, err := i.Pool.GetClient(ctx, k)
	if err != nil {
		return nil, err
	}

	rpc := client.RawConnection()

	return rpc, nil
}
