package main

import (
	"context"
	"flag"
	"fmt"
	"net"

	"go.uber.org/zap"

	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/core/io"
)

type MemStore struct {
}

func (m *MemStore) Get(ctx context.Context, req *pb.DcGetReq) (*pb.DcGetRes, error) {
	panic("implement me")
}

func (m *MemStore) Put(ctx context.Context, req *pb.DcPutReq) (*pb.DcRes, error) {
	panic("implement me")
}

func (m *MemStore) Remove(ctx context.Context, req *pb.DcRemoveReq) (*pb.DcRes, error) {
	panic("implement me")
}

func main() {
	var port = flag.Int("port", 9999, "port")
	core.ZapGlobalLevel(zap.InfoLevel)
	flag.Parse()

	store := &MemStore{}
	srv := io.GrpcServer()
	pb.RegisterDcRpcServer(srv, store)

	url := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", url)
	if err != nil {
		panic(err)
	}
	zap.S().Infow("Opened socket", "address", url)
	err = srv.Serve(lis)
	if err != nil {
		panic(err)
	}
}
