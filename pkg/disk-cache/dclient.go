package diskcache

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"github.com/anujga/dstk/pkg/ss"
	"github.com/anujga/dstk/pkg/ss/partrpc"
	"google.golang.org/grpc"
)

//Deprecated: Use the direct rpc interface
type Client interface {
	Get(key core.KeyT) ([]byte, error)
	Put(key, value core.KeyT, ttlSeconds float32) error
	Remove(key core.KeyT) error
}

// https://github.com/anujga/dstk/issues/40
type fwdClient struct {
	ss.ClientBase
}

func (i *fwdClient) Get(ctx context.Context, in *pb.DcGetReq, opts ...grpc.CallOption) (*pb.DcGetRes, error) {
	out := new(pb.DcGetRes)

	err := i.Fwd(ctx,
		in.Key,
		in,
		out,
		"/dstk.DcRpc/Get",
		opts...)

	return out, err

}

func (i *fwdClient) Put(ctx context.Context, in *pb.DcPutReq, opts ...grpc.CallOption) (*pb.DcRes, error) {
	out := new(pb.DcRes)

	err := i.Fwd(ctx,
		in.Key,
		in,
		out,
		"/dstk.DcRpc/Put",
		opts...)

	return out, err
}

func (i *fwdClient) Remove(ctx context.Context, in *pb.DcRemoveReq, opts ...grpc.CallOption) (*pb.DcRes, error) {
	out := new(pb.DcRes)

	err := i.Fwd(ctx,
		in.Key,
		in,
		out,
		"/dstk.DcRpc/Remove",
		opts...)

	return out, err
}

func NewClient(ctx context.Context, clientId string, seUrl string, opts ...grpc.DialOption) (pb.DcRpcClient, error) {
	seClient, err := se.NewSeClient(ctx, seUrl, opts...)
	if err != nil {
		return nil, err
	}

	clientPool, _ := partrpc.NewPartitionClientPool(
		clientId,
		dcClientFactory,
		seClient,
		opts...)

	return &fwdClient{
		ClientBase: ss.ClientBase{Pool: clientPool},
	}, nil
}

func dcClientFactory(conn *grpc.ClientConn) interface{} {
	return pb.NewDcRpcClient(conn)
}
