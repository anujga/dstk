package dc

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

	responses, err := d.reqHandler.HandleBlocking(req)
	if err != nil {
		return nil, err.Err()
	}

	res := responses.(*pb.DcDocument)
	return &pb.DcGetRes{
		Key:   req.Key(),
		Document: res,
	}, nil
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
	_, err := d.reqHandler.HandleBlocking(req)
	if err != nil {
		return nil, err.Err()
	}

	return &pb.DcRes{}, nil
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
	_, err := d.reqHandler.HandleBlocking(req)
	if err != nil {
		return nil, err.Err()
	}

	return &pb.DcRes{}, nil
}

func MakeServer(rh *ss.MsgHandler, resBufSize int64) *DiskCacheServer {
	return &DiskCacheServer{reqHandler: rh, log: zap.L(), resBufSize: resBufSize}
}
