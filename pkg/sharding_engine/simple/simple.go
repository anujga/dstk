package simple

import (
	"fmt"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/core/io"
	"google.golang.org/grpc"
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

	sock := io.GrpcServer()
	pb.RegisterSeWorkerApiServer(sock, server)
	pb.RegisterSeClientApiServer(sock, server)

	f := core.RunAsync(func() error {
		return sock.Serve(lis)
	})

	return f, sock, nil
}
