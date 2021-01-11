package main

import (
	"bytes"
	"context"
	"flag"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/core/io"
)

func main() {
	var url = flag.String("url", "0.0.0.0:9999", "0.0.0.0:9999")
	core.ZapGlobalLevel(zap.InfoLevel)
	flag.Parse()

	conn, err := grpc.DialContext(context.TODO(), *url, io.DefaultClientOpts()...)
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()
	rpc := pb.NewDcRpcClient(conn)

	k := []byte{65}
	v := []byte{66}

	_, err = rpc.Put(ctx, &pb.DcPutReq{
		Key:        k,
		Value:      v,
		TtlSeconds: 0,
	})

	if err != nil {
		panic(err)
	}

	got, err := rpc.Get(ctx, &pb.DcGetReq{Key: k})

	if bytes.Compare(v, got.Value) != 0 {
		panic(err)
	}
}
