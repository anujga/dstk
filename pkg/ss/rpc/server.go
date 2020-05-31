package main

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", ":9099")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pserver, err := NewPartServer(zap.S())
	pb.RegisterPartitionRpcServer(s, pserver)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
