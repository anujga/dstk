package main

import (
	"context"
	"fmt"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"go.uber.org/zap"
	"time"
)

var charset = []byte("abcdefghijklmnopqrstuvwxyz")

type CounterServer struct {
	reqHandler *ReqHandler
	resBufSize int64
	log        *zap.Logger
}

func (c *CounterServer) Remove(ctx context.Context, rpcReq *pb.CounterRemoveReq) (*pb.CounterRemoveRes, error) {
	ch := make(chan interface{}, c.resBufSize)
	req := newRemoveRequest(rpcReq.Key, ch)
	var exCode pb.Ex_ExCode
	var response interface{}
	var err error
	if response, err = c.reqHandler.handle(req); err != nil {
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
	if response, err := c.reqHandler.handle(req); err != nil {
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
	nano := time.Now().UnixNano()
	ch := make(chan interface{}, c.resBufSize)
	keyStr := string(charset[nano%26])
	if nano&1 == 1 {
		keyStr = keyStr + string(charset[13+time.Now().UnixNano()%13])
	}
	req := newIncRequest(rpcReq.Key, rpcReq.Value, float64(rpcReq.TtlSeconds), ch)
	var exCode pb.Ex_ExCode
	var response interface{}
	var err error
	if response, err = c.reqHandler.handle(req); err != nil {
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

func MakeServer(rh *ReqHandler, log *zap.Logger, resBufSize int64) *CounterServer {
	return &CounterServer{reqHandler: rh, log: log, resBufSize: resBufSize}
}
