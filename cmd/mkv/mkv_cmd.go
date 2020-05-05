package main

import (
	"flag"
	"fmt"
	pb "github.com/anujga/dstk/build/gen"
	"github.com/anujga/dstk/pkg/mkv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"os"
)

func main() {
	var port = flag.Int("port", 6001, "grpc port")
	flag.Parse()
	log, err := zap.NewProduction()
	if err != nil {
		println("Failed to open logger %s", err)
		os.Exit(-1)
	}
	slog := log.Sugar()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		slog.Fatalw("failed to listen",
			"port", port,
			"err", err)
	}
	grpcServer := grpc.NewServer()
	s, err := mkv.MakeServer(int32(*port))
	if err != nil {
		slog.Fatalw("failed to initialize server object",
			"port", port,
			"err", err)
	}

	pb.RegisterMkvServer(grpcServer, s)

	if err = grpcServer.Serve(lis); err != nil {
		slog.Fatalw("failed to start server",
			"port", port,
			"err", err)
		os.Exit(-2)
	}

}
