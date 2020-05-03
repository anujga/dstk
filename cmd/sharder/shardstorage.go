package main

import (
	"context"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	port = ":50000"
)

type shardStoreUpdater struct {
	dstk.UnimplementedShardStorageServer
	store *ShardStore
}

func (s *shardStoreUpdater) CreatePartition(_ context.Context, in *dstk.CreateReq) (*dstk.ChangeRes, error) {
	partition := in.GetC()
	logger := zap.L()
	logger.Info("To Create: ", zap.Any("partition", partition))
	if partition == nil {
		ex := core.NewErr(core.ErrInvalidPartition, "Invalid Input Partition")
		logger.Error("Invalid Input Partition: ", zap.Any("Error", ex))
		res := dstk.ChangeRes{Ex: &ex.Ex, Success: false}
		return &res, nil
	}
	s.store.Create(partition)
	res := dstk.ChangeRes{Success: false}
	return &res, nil
}

func (s *shardStoreUpdater) SplitPartition(_ context.Context, in *dstk.SplitReq) (*dstk.ChangeRes, error) {
	logger := zap.L()
	c := in.GetC()
	n1 := in.GetN1()
	n2 := in.GetN2()
	if c == nil || n1 == nil || n2 == nil {
		ex := core.NewErr(core.ErrInvalidPartition, "Invalid Input Partition")
		logger.Error("Invalid Input Partition: ", zap.Any("Error", ex))
		res := dstk.ChangeRes{Ex: &ex.Ex, Success: false}
		return &res, nil
	}
	s.store.Split(c, n1, n2)
	res := dstk.ChangeRes{Success: false}
	return &res, nil
}

func (s *shardStoreUpdater) MergePartition(_ context.Context, in *dstk.MergeReq) (*dstk.ChangeRes, error) {
	logger := zap.L()
	c1 := in.GetC1()
	c2 := in.GetC2()
	n := in.GetN()
	if c1 == nil || c2 == nil || n == nil {
		ex := core.NewErr(core.ErrInvalidPartition, "Invalid Input Partition")
		logger.Error("Invalid Input Partition: ", zap.Any("Error", ex))
		res := dstk.ChangeRes{Ex: &ex.Ex, Success: false}
		return &res, nil
	}
	s.store.Merge(c1, c2, n)
	res := dstk.ChangeRes{Success: false}
	return &res, nil
}

func (s *shardStoreUpdater) FindPartition(_ context.Context, in *dstk.Find_Req) (*dstk.Find_Res, error) {
	key := in.GetKey()
	logger := zap.L()
	logger.Info("Find Partition for: ", zap.Any("key", key))
	partition := s.store.Find(key)
	res := dstk.Find_Res{Par: &partition}
	return &res, nil
}

func initLogger() {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
}

func createServer() {
	logger := zap.L()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("failed to listen: ", zap.Error(err))
	}
	s := grpc.NewServer()
	ssu := shardStoreUpdater{store: NewShardStore()}
	dstk.RegisterShardStorageServer(s, &ssu)
	logger.Info("Started Server ", zap.Any("Server:", s.GetServiceInfo()))
	logger.Info("Waiting for assignments")
	if err := s.Serve(lis); err != nil {
		logger.Error("failed to serve: ", zap.Error(err))
	}
}

func main() {
	initLogger()
	logger := zap.L()
	defer syncLogs(logger)
	// will return only when there is an error serving or graceful shutdown happens
	createServer()
	logger.Info("Shutting down the gRPC shardStoreUpdater")
}

func syncLogs(logger *zap.Logger) {
	err := logger.Sync()
	if err != nil {
		log.Fatal("Error syncing log!", err)
	}
}
