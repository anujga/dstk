package mkv

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/sharding_engine"
	"google.golang.org/grpc"
)

//This is the interface used by clients of mkv
type Client interface {
	Get(ctx context.Context, key []byte) ([]byte, error)
}

// todo: explore the usage of envoy instead of manually creating grpc conn_pool
// for stats, rate limiting, retry, round robin, proximity routing, auth ...
// istio with headless service should make this pretty trivial
// https://github.com/istio/istio/issues/10659
type mkvClient struct {
	slice se.ThickClient
	pool  core.ConnPool
}

func ShardedClient(slice se.ThickClient, opts ...grpc.DialOption) Client {
	return &mkvClient{
		slice: slice,
		pool: core.NonExpiryPool(&rpcConnFactory{
			opts: opts,
		}),
	}
}

func (m *mkvClient) Get(ctx context.Context, key []byte) ([]byte, error) {
	part, err := m.slice.Get(context.TODO(), key)
	if err != nil {
		return nil, err
	}

	url := part.GetUrl()
	o, err := m.pool.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	conn := o.(*unitPartition)
	value, err := conn.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// todo: extract this boilerplate into core

type rpcConnFactory struct {
	opts []grpc.DialOption
}

func (c *rpcConnFactory) Open(ctx context.Context, url string) (interface{}, error) {
	conn, err := grpc.DialContext(ctx, url, c.opts...)
	if err != nil {
		return nil, err
	}
	client := pb.NewMkvClient(conn)
	return &unitPartition{conn: conn, client: client, url: url}, nil
}

func (c *rpcConnFactory) Close(conn interface{}) error {
	conn2 := conn.(*unitPartition)
	return conn2.Close()
}

// makes call to a single partition. it may choose to connect with more than
// one node in case of master slave and send request in round robin.
type unitPartition struct {
	// required only for close
	conn *grpc.ClientConn

	// mostly used for debugging
	url string

	client pb.MkvClient
}

func (s *unitPartition) Get(c context.Context, key []byte) ([]byte, error) {
	r, err := s.client.Get(c, &pb.GetReq{
		Key:         key,
		PartitionId: 0,
	})

	if err != nil {
		return nil, err
	}

	return r.Payload, nil
}

func (s *unitPartition) Close() error {
	return s.conn.Close()
}
