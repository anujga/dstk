package dc

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core/control"
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
	req := &DcRequest{
		grpcRequest: rpcReq,
		C:           d.resposeBuffer(),
		Ctx:         ctx,
	}

	res := d.reqHandler.HandleBlocking(req)
	if res.Err != nil {
		return nil, res.Err.Err()
	}

	rs := res.Res.([]byte)
	return &pb.DcGetRes{
		Key:   req.Key(),
		Value: rs,
	}, nil
}

func (d *DiskCacheServer) Put(ctx context.Context, rpcReq *pb.DcPutReq) (*pb.DcRes, error) {
	if rpcReq.GetKey() == nil || rpcReq.GetValue() == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"Key/value cannot be null")
	}
	req := &DcRequest{
		grpcRequest: rpcReq,
		C:           d.resposeBuffer(),
		Ctx:         ctx,
	}

	r := d.reqHandler.HandleBlocking(req)
	if r.Err != nil {
		return nil, r.Err.Err()
	}

	return &pb.DcRes{}, nil
}

func (d *DiskCacheServer) resposeBuffer() chan *control.Response {
	return make(chan *control.Response, d.resBufSize)
}

func (d *DiskCacheServer) Remove(ctx context.Context, rpcReq *pb.DcRemoveReq) (*pb.DcRes, error) {
	if rpcReq.GetKey() == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"Key cannot be null")
	}
	req := &DcRequest{
		grpcRequest: rpcReq,
		C:           d.resposeBuffer(),
	}
	r := d.reqHandler.HandleBlocking(req)
	if r.Err != nil {
		return nil, r.Err.Err()
	}

	return &pb.DcRes{}, nil
}

func MakeServer(rh *ss.MsgHandler, resBufSize int64) *DiskCacheServer {
	return &DiskCacheServer{reqHandler: rh, log: zap.L(), resBufSize: resBufSize}
}
