package main

import (
	"context"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"go.uber.org/zap"
)

type PartServer struct {
	log *zap.Logger
	slog *zap.SugaredLogger
}

func (p *PartServer) OnMessage(ctx context.Context, msg *dstk.PartMsg) (*dstk.PartResponse, error) {
	p.slog.Infow("Received", "msg", msg)
	return &dstk.PartResponse{}, nil
}

func NewPartServer(logger *zap.SugaredLogger) (*PartServer, error)  {
	return &PartServer{slog: logger}, nil
}
