package ss

import (
	"context"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"github.com/anujga/dstk/pkg/ss/common"
	"github.com/anujga/dstk/pkg/ss/node"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type WorkerGrpcServer struct {
	Server     *grpc.Server
	psyncer    *node.PartsSyncer
	MsgHandler *MsgHandler
	logger     *zap.Logger
}

func (wgs *WorkerGrpcServer) Start(network, address string) error {
	wgs.MsgHandler.w.Start()
	wgs.psyncer.Start()
	reflection.Register(wgs.Server)
	if lis, err := net.Listen(network, address); err == nil {
		zap.S().Infow("Opened socket", "address", address)
		return wgs.Server.Serve(lis)
	} else {
		return err
	}
}

func NewWorkerServer(seUrl string, wid se.WorkerId, consumerFactory common.ConsumerFactory) (*WorkerGrpcServer, error) {
	logger := zap.L()
	wa, err2 := node.NewActor(consumerFactory, wid)
	if err2 != nil {
		return nil, err2
	}

	seClient, err := se.NewSeWorker(context.TODO(), seUrl, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	grpc_prometheus.EnableHandlingTimeHistogram()
	//https://github.com/grpc-ecosystem/go-grpc-middleware
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			//grpc_ctxtags.StreamServerInterceptor(),
			//grpc_opentracing.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			//grpc_ctxtags.UnaryServerInterceptor(),
			//grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		)),
	)
	return &WorkerGrpcServer{
		Server:     s,
		MsgHandler: &MsgHandler{wa},
		logger:     logger,
		psyncer:    node.NewSyncer(wa, seClient),
	}, nil
}
