package mkv

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/slicer"
	"google.golang.org/grpc"
)

//todo: explore the usage of envoy instead of manually creating grpc connections
//for stats, rate limiting, retry, round robin / local zone, auth ...

type Client interface {
	Get(key []byte) ([]byte, error)
}

type staticClient struct {
	conn   *grpc.ClientConn
	client pb.MkvClient
}

func newRpcClient(serverUrl string) (*staticClient, error) {
	conn, err := grpc.Dial(serverUrl)
	if err != nil {
		return nil, err
	}
	client := pb.NewMkvClient(conn)
	return &staticClient{conn: conn, client: client}, nil
}

func (s *staticClient) Get(key []byte) ([]byte, error) {
	r, err := s.client.Get(context.TODO(), &pb.GetReq{
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
	slice slicer.SliceRdr
	pool  core.ConnPool
}

func (m *mkvClient) Get(key []byte) ([]byte, error) {
	part, err := m.slice.Get(key)
	if err != nil {
		return nil, err
	}

	url := part.Url()
	o, err := m.pool.Get(url)
	if err != nil {
		return nil, err
	}

	conn := o.(*staticClient)
	value, err := conn.Get(key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func UsingSlicer(slice slicer.SliceRdr) Client {
	return &mkvClient{
		slice: slice,
		pool:  core.NonExpiryPool(&rpcConnFactory{}),
	}
}

type rpcConnFactory struct {
	//auth string
}

func (c *rpcConnFactory) Open(url string) (interface{}, error) {
	return newRpcClient(url)
}

func (c *rpcConnFactory) Close(conn interface{}) error {
	conn2 := conn.(*staticClient)
	return conn2.Close()
}
