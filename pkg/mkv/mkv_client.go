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

type staticClient struct {
	//todo: explore the usage of envoy instead of manually creating grpc connections
	//for stats, rate limiting, retry, round robin / local zone, auth ...
	conn *grpc.ClientConn

	client pb.MkvClient
}

func (s *staticClient) Get(c context.Context, key []byte) ([]byte, error) {
	r, err := s.client.Get(c, &pb.GetReq{
		Key:         key,
		PartitionId: 0,
	})

	if err != nil {
		return nil, err
	}

	if r.Ex.GetId() != pb.Ex_SUCCESS {
		return nil, core.WrapEx(r.Ex)
	}

	return r.Payload, nil
}

func (s *staticClient) Close() error {
	return s.conn.Close()
}

type mkvClient struct {
	slice se.ThickClient
	pool  core.ConnPool
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

	conn := o.(*staticClient)
	value, err := conn.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func UsingSlicer(slice se.ThickClient) Client {
	return &mkvClient{
		slice: slice,
		pool:  core.NonExpiryPool(&rpcConnFactory{}),
	}
}

type rpcConnFactory struct {
	//auth string
}

func (c *rpcConnFactory) Open(ctx context.Context, url string) (interface{}, error) {
	conn, err := grpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	client := pb.NewMkvClient(conn)
	return &staticClient{conn: conn, client: client}, nil
}

func (c *rpcConnFactory) Close(conn interface{}) error {
	conn2 := conn.(*staticClient)
	return conn2.Close()
}
