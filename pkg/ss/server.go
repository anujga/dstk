package ss

import (
	"context"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"net"
)

type WorkerCtrlServerImpl struct {
	MsgHandler *MsgHandler
	logger     *zap.Logger
}

func (w *WorkerCtrlServerImpl) SplitPartition(ctx context.Context, req *dstk.SplitPartReq) (*dstk.SplitPartResponse, error) {
	cm := &CtrlMsg{
		grpcReq:   req,
		ctx:       ctx,
		ch: make(chan interface{}, 0),
	}
	if responses, err := w.MsgHandler.Handle(cm); err != nil {
		return nil, err
	} else {
		if len(responses) == 0 {
			return &dstk.SplitPartResponse{}, nil
		} else {
			w.logger.Error("invalid response", zap.Any("responses", responses))
			return nil, status.Error(codes.Internal, "internal")
		}
	}
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
	s := grpc.NewServer()
	dstk.RegisterWorkerCtrlServer(s, ws)
	return &WorkerGrpcServer{
		Server:     s,
		MsgHandler: mh,
		logger:     logger,
	}, nil
}