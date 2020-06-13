package diskcache

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"google.golang.org/grpc"
	"time"
)

type Client interface {
	Get(key core.KeyT) ([]byte, error)
	Put(key, value core.KeyT) error
	Remove(key core.KeyT) error
}

type impl struct {
	tc     se.ThickClient
	rpcMap *core.ConcurrentMap
}

func newClientLambda(ctx context.Context) func(target interface{}) (interface{}, error) {
	return func(target interface{}) (interface{}, error) {
		conn, err := grpc.DialContext(ctx, target.(string), grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		return pb.NewDcRpcClient(conn), err
	}
}

func (i *impl) getRpc(ctx context.Context, key core.KeyT) (pb.DcRpcClient, error) {
	if part, err := i.tc.Get(ctx, key); err == nil {
		// TODO will taking a lock on map become a choke point?
		if res, err := i.rpcMap.ComputeIfAbsent(part.GetUrl(), newClientLambda(ctx)); err == nil {
			return res.(pb.DcRpcClient), nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

// thread safe
func (i *impl) Get(key core.KeyT) ([]byte, error) {
	ctx := context.TODO()
	if rpc, err := i.getRpc(ctx, key); err == nil {
		rpcResponse, err := rpc.Get(ctx, &pb.DcGetReq{Key: key})
		if err != nil {
			return nil, err
		} else {
			return rpcResponse.GetValue(), nil
		}
	} else {
		return nil, err
	}
}

func (i *impl) Put(key, value core.KeyT) error {
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
	c := impl{
		tc:     se.NewThickClient("c1", seClient),
		rpcMap: core.NewConcurrentMap(),
	}
	// thick client state is not loaded immediately, so waiting. we need to fix this
	time.Sleep(time.Second * 80)
	return &c, nil
}
