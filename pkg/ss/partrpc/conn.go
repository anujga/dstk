package partrpc

import (
	"context"
	"github.com/anujga/dstk/pkg/core"
	"google.golang.org/grpc"
)

type RpcClientFactory = func(*grpc.ClientConn) interface{}

type rpcConnFactory struct {
	opts             []grpc.DialOption
	rpcClientFactory RpcClientFactory
}

func (scf *rpcConnFactory) Open(ctx context.Context, url string) (interface{}, error) {
	conn, err := grpc.DialContext(ctx, url, scf.opts...)
	if err != nil {
		return nil, err
	}
	return &partClient{
		grpcClient: scf.rpcClientFactory(conn),
		conn:       conn,
		url:        url,
	}, err
}

func (scf *rpcConnFactory) Close(i interface{}) error {
	return i.(PartitionClient).Close()
}

func newRpcConnFactory(rpcClientFactory RpcClientFactory, opts ...grpc.DialOption) core.ConnectionFactory {
	return &rpcConnFactory{rpcClientFactory: rpcClientFactory, opts: opts}
}
