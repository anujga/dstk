package gateway

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/core/io"
	diskcache "github.com/anujga/dstk/pkg/disk-cache"
	"google.golang.org/grpc"
	"net"
)

type Config struct {
	Url      string
	SeUrl    string
	ClientId string
}

func startServer(url string, server pb.DcRpcServer) (*core.FutureErr, *grpc.Server, error) {
	lis, err := net.Listen("tcp", url)
	if err != nil {
		return nil, nil, err
	}

	sock := io.GrpcServer()

	pb.RegisterDcRpcServer(sock, server)

	f := core.RunAsync(func() error {
		return sock.Serve(lis)
	})

	return f, sock, nil
}

type fwdProxy struct {
	rpc  pb.DcRpcClient
	opts []grpc.CallOption
}

func (p *fwdProxy) Get(c context.Context, in *pb.DcGetReq) (*pb.DcGetRes, error) {
	return p.rpc.Get(c, in, p.opts...)
}

func (p *fwdProxy) Put(c context.Context, in *pb.DcPutReq) (*pb.DcRes, error) {
	return p.rpc.Put(c, in, p.opts...)

}
func (p *fwdProxy) Remove(c context.Context, in *pb.DcRemoveReq) (*pb.DcRes, error) {
	return p.rpc.Remove(c, in, p.opts...)
}

func GatewayMode(c *Config) (*core.FutureErr, error) {

	opts := io.DefaultClientOpts()
	rpc, err := diskcache.NewClient(
		context.TODO(),
		c.ClientId,
		c.SeUrl,
		opts...)

	if err != nil {
		return nil, err
	}

	proxy := &fwdProxy{
		rpc:  rpc,
		opts: []grpc.CallOption{},
	}

	f, _, err := startServer(c.Url, proxy)
	if err != nil {
		return nil, err
	}

	return f, nil
}
