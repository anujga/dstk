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

func (d *DiskCacheServer) Get(ctx context.Context, rpcReq *pb.DcGetReq) (*pb.DcGetRes, error) {
	if rpcReq.GetKey() == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"Key cannot be null")
	}
	ch := make(chan interface{}, d.resBufSize)
	req := &DcRequest{
		grpcRequest: rpcReq,
		C:           ch,
	}
	if responses, err := d.reqHandler.Handle(req); err != nil {
		return nil, err
	} else {
		res := responses[0]
		switch res.(type) {
		case error:
			return nil, res.(error)
		case []byte:
			return &pb.DcGetRes{
				Value: responses[0].([]byte),
			}, err
		default:
			d.log.Error("invalid response", zap.Any("response", res))
			return nil, status.Error(codes.Internal, "internal")
		}
	}
}

func (d *DiskCacheServer) Put(ctx context.Context, rpcReq *pb.DcPutReq) (*pb.DcRes, error) {
	if rpcReq.GetKey() == nil || rpcReq.GetValue() == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"Key/value cannot be null")
	}
	ch := make(chan interface{}, d.resBufSize)
	req := &DcRequest{
		grpcRequest: rpcReq,
		C:           ch,
		Ctx:         ctx,
	}
	if responses, err := d.reqHandler.Handle(req); err != nil {
		return nil, err
	} else {
		if len(responses) == 0 {
			return &pb.DcRes{}, nil
		} else {
			d.log.Error("invalid response", zap.Any("responses", responses))
			return nil, status.Error(codes.Internal, "internal")
		}
	}
}

func (d *DiskCacheServer) Remove(ctx context.Context, rpcReq *pb.DcRemoveReq) (*pb.DcRes, error) {
	if rpcReq.GetKey() == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"Key cannot be null")
	}
	ch := make(chan interface{}, d.resBufSize)
	req := &DcRequest{
		grpcRequest: rpcReq,
		C:           ch,
	}
	if responses, err := d.reqHandler.Handle(req); err != nil {
		return nil, err
	} else {
		if len(responses) == 0 {
			return &pb.DcRes{}, nil
		} else {
			d.log.Error("invalid response", zap.Any("responses", responses))
			return nil, status.Error(codes.Internal, "internal")
		}
	}
}

func MakeServer(rh *ss.MsgHandler, resBufSize int64) *DiskCacheServer {
	return &DiskCacheServer{reqHandler: rh, log: zap.L(), resBufSize: resBufSize}
}
