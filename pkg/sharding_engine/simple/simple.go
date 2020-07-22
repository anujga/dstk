package simple

import (
	"fmt"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type WorkerAndClient interface {
	pb.SeWorkerApiServer
	pb.SeClientApiServer
}

func StartServer(port int, server WorkerAndClient) (*core.FutureErr, *grpc.Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, nil, err
	}

	sock := grpc.NewServer()
	reflection.Register(sock)
	pb.RegisterSeWorkerApiServer(sock, server)
	pb.RegisterSeClientApiServer(sock, server)

	f := core.RunAsync(func() error {
		return sock.Serve(lis)
	})

	return f, sock, nil
}
