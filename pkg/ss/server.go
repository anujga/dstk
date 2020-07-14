package ss

import (
	"context"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type WorkerCtrlServerImpl struct {
	MsgHandler *MsgHandler
	logger     *zap.Logger
}

func (w *WorkerCtrlServerImpl) SplitPartition(ctx context.Context, req *dstk.SplitPartReq) (*dstk.SplitPartResponse, error) {
	cm := &CtrlMsg{
		grpcReq: req,
		ctx:     ctx,
		ch:      make(chan interface{}, 0),
	}
	_, err := w.MsgHandler.HandleBlocking(cm)
	if err != nil {
		return nil, err.Err()
	}

	return &dstk.SplitPartResponse{}, nil
}

type WorkerGrpcServer struct {
	Server     *grpc.Server
	MsgHandler *MsgHandler
	logger     *zap.Logger
}

func (wgs *WorkerGrpcServer) Start(network, address string) error {
	wgs.MsgHandler.w.Start()
	reflection.Register(wgs.Server)
	if lis, err := net.Listen(network, address); err == nil {
		return wgs.Server.Serve(lis)
	} else {
		return err
	}
}

func NewWorkerServer(seUrl string, wid se.WorkerId, consumerFactory ConsumerFactory, initStateMaker func() interface{}) (*WorkerGrpcServer, error) {
	logger := zap.L()
	seClient, err := se.NewSeWorker(context.TODO(), seUrl, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	wa := NewPartitionMgr2(wid, consumerFactory, seClient, initStateMaker)
	mh := &MsgHandler{wa}
	ws := &WorkerCtrlServerImpl{
		MsgHandler: mh,
		logger:     logger,
	}

	//s := grpc.NewServer()

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

	dstk.RegisterWorkerCtrlServer(s, ws)

	return &WorkerGrpcServer{
		Server:     s,
		MsgHandler: mh,
		logger:     logger,
	}, nil
}
