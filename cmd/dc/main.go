package main

import (
	"flag"
	"fmt"
	"net"

	"go.uber.org/zap"

	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/core/io"
)

func main() {
	var port = flag.Int("port", 9999, "port")
	var path = flag.String("data", "/data/", "data")
	core.ZapGlobalLevel(zap.InfoLevel)
	flag.Parse()

	store, err := NewbadgerStore(*path)
	if err != nil {
		panic(err)
	}
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
