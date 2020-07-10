package ss

import (
	"context"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type WorkerCtrlServerImpl struct {
	logger *zap.Logger
}

func (w *WorkerCtrlServerImpl) SplitPartition(ctx context.Context, req *dstk.SplitPartReq) (*dstk.SplitPartResponse, error) {
	panic("implement me")
}

type WorkerGrpcServer struct {
	Server     *grpc.Server
	MsgHandler *MsgHandler
	logger *zap.Logger
}

func (wgs *WorkerGrpcServer) Start(network, address string) error {
	wgs.MsgHandler.Start()
	reflection.Register(wgs.Server)
	lis, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	return wgs.Server.Serve(lis)
}

func NewWorkerServer(seUrl string, wid se.WorkerId, consumerFactory ConsumerFactory, logger *zap.Logger) (*WorkerGrpcServer, error) {
	s := grpc.NewServer()
	ws := &WorkerCtrlServerImpl{
		logger: logger,
	}
	dstk.RegisterWorkerCtrlServer(s, ws)
	seClient, err := se.NewSeWorker(context.TODO(), seUrl, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	wa := NewPartitionMgr2(wid, consumerFactory, seClient, func() interface{} {
		return nil
	})
	return &WorkerGrpcServer{
		Server:     s,
		MsgHandler: &MsgHandler{wa},
		logger: logger,
	}, nil
}
