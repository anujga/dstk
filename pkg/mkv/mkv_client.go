package mkv

import (
	"context"
	pb "github.com/anujga/dstk/build/gen"
	"github.com/anujga/dstk/pkg/core"
	"google.golang.org/grpc"
)

type Client interface {
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
