package main

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DiskCacheServer struct {
	reqHandler *ss.MsgHandler
	resBufSize int64
	log        *zap.Logger
}

func (d DiskCacheServer) Get(ctx context.Context, rpcReq *pb.DcGetReq) (*pb.DcGetRes, error) {
	ch := make(chan interface{}, d.resBufSize)
	req := &DcRequest{
		grpcRequest: rpcReq,
		C:           ch,
	}

	//if req.Key() == nil {
	//	req.grpcRequest.(*pb.DcGetReq).Key = []byte("asd")
	//}

	if req.Key() == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"Key cannot be null")
	}

	if response, err := d.reqHandler.Handle(req); err != nil {
		return &pb.DcGetRes{
			Value: nil,
		}, err
	} else {
		return &pb.DcGetRes{
			Value: response.([]byte),
		}, err
	}
}

func (d DiskCacheServer) Put(ctx context.Context, req *pb.DcPutReq) (*pb.DcRes, error) {
	panic("implement me")
}

func (d DiskCacheServer) Remove(ctx context.Context, req *pb.DcRemoveReq) (*pb.DcRes, error) {
	panic("implement me")
}

func MakeServer(rh *ss.MsgHandler, log *zap.Logger, resBufSize int64) *DiskCacheServer {
	return &DiskCacheServer{reqHandler: rh, log: log, resBufSize: resBufSize}
}
