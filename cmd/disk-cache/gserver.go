package main

import (
	"context"
	"fmt"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss"
	"go.uber.org/zap"
)

type DiskCacheServer struct {
	reqHandler *ss.MsgHandler
	resBufSize int64
	log        *zap.Logger
}

func (d DiskCacheServer) Get(ctx context.Context, rpcReq *pb.DcGetReq) (*pb.DcGetRes, error) {
	ch := make(chan interface{}, d.resBufSize)
	req := newGetRequest(rpcReq.GetKey(), ch)
	if response, err := d.reqHandler.Handle(req); err != nil {
		d.log.Error("Request handling  failed",
			zap.String("req", fmt.Sprintf("%v", rpcReq)), zap.Error(err))
		ex := &pb.Ex{
			Id:  pb.Ex_ERR_UNSPECIFIED,
			Msg: "internal error",
		}
		return &pb.DcGetRes{
			Ex:    ex,
			Value: nil,
		}, err
	} else {
		ex := &pb.Ex{
			Id: pb.Ex_SUCCESS,
		}
		return &pb.DcGetRes{
			Ex:    ex,
			Value: response.([]byte),
		}, err
	}
}

func (d DiskCacheServer) Put(ctx context.Context, req *pb.DcPutReq) (*pb.DcPutRes, error) {
	panic("implement me")
}

func (d DiskCacheServer) Remove(ctx context.Context, req *pb.DcRemoveReq) (*pb.DcRemoveRes, error) {
	panic("implement me")
}

func MakeServer(rh *ss.MsgHandler, log *zap.Logger, resBufSize int64) *DiskCacheServer {
	return &DiskCacheServer{reqHandler: rh, log: log, resBufSize: resBufSize}
}
