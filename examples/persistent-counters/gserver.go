package main

import (
	"context"
	"fmt"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss"
	"go.uber.org/zap"
)

type CounterServer struct {
	reqHandler *ss.MsgHandler
	resBufSize int64
	log        *zap.Logger
}

func (c *CounterServer) Remove(ctx context.Context, rpcReq *pb.CounterRemoveReq) (*pb.CounterRemoveRes, error) {
	ch := make(chan interface{}, c.resBufSize)
	req := newRemoveRequest(rpcReq.Key, ch)
	var exCode pb.Ex_ExCode
	var response interface{}
	var err error
	if response, err = c.reqHandler.Handle(req); err != nil {
		c.log.Error("Request handling  failed",
			zap.String("req", fmt.Sprintf("%v", rpcReq)), zap.Error(err))
		exCode = pb.Ex_ERR_UNSPECIFIED
	} else {
		exCode = pb.Ex_SUCCESS
	}
	ex := pb.Ex{
		Id:  exCode,
		Msg: response.(string),
	}
	return &pb.CounterRemoveRes{Ex: &ex}, err
}

func (c *CounterServer) Get(ctx context.Context, rpcReq *pb.CounterGetReq) (*pb.CounterGetRes, error) {
	ch := make(chan interface{}, c.resBufSize)
	req := newGetRequest(rpcReq.Key, ch)
	if response, err := c.reqHandler.Handle(req); err != nil {
		c.log.Error("Request handling  failed",
			zap.String("req", fmt.Sprintf("%v", rpcReq)), zap.Error(err))
		ex := &pb.Ex{
			Id:  pb.Ex_ERR_UNSPECIFIED,
			Msg: "internal error",
		}
		return &pb.CounterGetRes{
			Ex:    ex,
			Value: 0,
		}, err
	} else {
		ex := &pb.Ex{
			Id: pb.Ex_SUCCESS,
		}
		return &pb.CounterGetRes{
			Ex:    ex,
			Value: response.(int64),
		}, err
	}
}

func (c *CounterServer) Inc(ctx context.Context, rpcReq *pb.CounterIncReq) (*pb.CounterIncRes, error) {
	ch := make(chan interface{}, c.resBufSize)
	req := newIncRequest(rpcReq.Key, rpcReq.Value, float64(rpcReq.TtlSeconds), ch)
	var exCode pb.Ex_ExCode
	var response interface{}
	var err error
	if response, err = c.reqHandler.Handle(req); err != nil {
		c.log.Error("Request handling  failed",
			zap.String("req", fmt.Sprintf("%v", rpcReq)), zap.Error(err))
		exCode = pb.Ex_ERR_UNSPECIFIED
	} else {
		exCode = pb.Ex_SUCCESS
	}
	ex := pb.Ex{
		Id:  exCode,
		Msg: response.(string),
	}
	return &pb.CounterIncRes{Ex: &ex}, err
}

func MakeServer(rh *ss.MsgHandler, log *zap.Logger, resBufSize int64) *CounterServer {
	return &CounterServer{reqHandler: rh, log: log, resBufSize: resBufSize}
}
