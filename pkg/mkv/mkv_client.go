package mkv

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"google.golang.org/grpc"
	"io"
)

//todo: explore the usage of envoy instead of manually creating grpc connections
//for stats, rate limiting, retry, round robin / local zone, auth ...

type Client interface {
	io.Closer
	Get(key []byte) ([]byte, error)
}

type staticClient struct {
	conn   *grpc.ClientConn
	client pb.MkvClient
}

func MakeClient(serverUrl string) (Client, error) {
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

type mkvClientFactory struct {
	auth string
}

func (c *mkvClientFactory) Open(url string) (interface{}, error) {
	return MakeClient(url)
}

func (c *mkvClientFactory) Close(conn interface{}) error {
	conn2 := conn.(Client)
	return conn2.Close()
}
