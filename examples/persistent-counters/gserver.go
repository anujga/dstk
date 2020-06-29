package main

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CounterServer struct {
	reqHandler *ss.MsgHandler
	resBufSize int64
	log        *zap.Logger
}

func (c *CounterServer) Remove(ctx context.Context, rpcReq *pb.CounterRemoveReq) (*pb.CounterRemoveRes, error) {
	ch := make(chan interface{}, c.resBufSize)
	req := newRemoveRequest(rpcReq.Key, ch)
	if responses, err := c.reqHandler.Handle(req); err != nil {
		return nil, err
	} else {
		if len(responses) == 0 {
			return &pb.CounterRemoveRes{}, nil
		} else {
			c.log.Error("invalid response", zap.Any("responses", responses))
			return nil, status.Error(codes.Internal, "internal")
		}
	}
}

func (c *CounterServer) Get(ctx context.Context, rpcReq *pb.CounterGetReq) (*pb.CounterGetRes, error) {
	ch := make(chan interface{}, c.resBufSize)
	req := newGetRequest(rpcReq.Key, ch)
	if responses, err := c.reqHandler.Handle(req); err != nil {
		return nil, err
	} else {
		res := responses[0]
		switch res.(type) {
		case error:
			return nil, res.(error)
		case int64:
			return &pb.CounterGetRes{
				Value: responses[0].(int64),
			}, err
		default:
			c.log.Error("invalid response", zap.Any("response", res))
			return nil, status.Error(codes.Internal, "internal")
		}
	}
}

func (c *CounterServer) Inc(ctx context.Context, rpcReq *pb.CounterIncReq) (*pb.CounterIncRes, error) {
	ch := make(chan interface{}, c.resBufSize)
	req := newIncRequest(rpcReq.Key, rpcReq.Value, float64(rpcReq.TtlSeconds), ch)
	if responses, err := c.reqHandler.Handle(req); err != nil {
		return nil, err
	} else {
		if len(responses) == 0 {
			return &pb.CounterIncRes{}, nil
		} else {
			c.log.Error("invalid response", zap.Any("responses", responses))
			return nil, status.Error(codes.Internal, "internal")
		}
	}
}

func MakeServer(rh *ss.MsgHandler, log *zap.Logger, resBufSize int64) *CounterServer {
	return &CounterServer{reqHandler: rh, log: log, resBufSize: resBufSize}
}
