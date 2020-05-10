package main

import (
	"context"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
)

const (
	port = ":50000"
)

type shardStoreUpdater struct {
	dstk.UnimplementedShardStorageServer
	store *ShardStore
}

func (s *shardStoreUpdater) CreatePartition(_ context.Context, in *dstk.CreateJobPartReq) (*dstk.ChangeRes, error) {
	logger := zap.L()
	markings := in.GetMarkings()
	if markings == nil || len(markings) < 1 {
		ex := core.NewErr(dstk.Ex_INVALID_ARGUMENT, "Invalid markings")
		logger.Error("Invalid markings: ", zap.Any("Error", ex))
		res := dstk.ChangeRes{Ex: ex.Ex, Success: false}
		return &res, nil
	}
	old := []byte(nil)
	var partitions []*dstk.Partition
	for _, cur := range markings {
		part := dstk.Partition{Id: generatePartitionId(), Start: old, End: cur, Url: getUrl()}
		partitions = append(partitions, &part)
		old = cur
	}
	lastPart := dstk.Partition{Id: generatePartitionId(), Start: markings[len(markings) - 1], Url: getUrl()}
	err := s.store.Create(in.GetJobId(), partitions, &lastPart)
 	if err != nil {
		ex := core.NewErr(dstk.Ex_CONFLICT, err.Error())
		logger.Error("CreatePartition Error: ", zap.Any("Error", ex))
		res := dstk.ChangeRes{Ex: ex.Ex, Success: false}
		return &res, nil
	}
	res := dstk.ChangeRes{Success: false}
	return &res, nil
}

func (s *shardStoreUpdater) SplitPartition(_ context.Context, in *dstk.SplitReq) (*dstk.ChangeRes, error) {
	logger := zap.L()
	jobId := in.GetJobId()
	marking := in.GetMarking()
	if jobId == 0 || marking == nil {
		ex := core.NewErr(dstk.Ex_INVALID_ARGUMENT, "Invalid job id or marking")
		logger.Error("Invalid Marking: ", zap.Any("Error", ex))
		res := dstk.ChangeRes{Ex: ex.Ex, Success: false}
		return &res, nil
	}
	err := s.store.Split(jobId, marking)
	if err != nil {
		ex := core.NewErr(dstk.Ex_BAD_PARTITION, err.Error())
		logger.Error("SplitPartition Error: ", zap.Any("Error", ex))
		res := dstk.ChangeRes{Ex: ex.Ex, Success: false}
		return &res, nil
	}
	res := dstk.ChangeRes{Success: true}
	return &res, nil
}

func (s *shardStoreUpdater) MergePartition(_ context.Context, in *dstk.MergeReq) (*dstk.ChangeRes, error) {
	logger := zap.L()
	c1 := in.GetC1()
	c2 := in.GetC2()
	n := in.GetN()
	if c1 == nil || c2 == nil || n == nil {
		ex := core.NewErr(dstk.Ex_INVALID_ARGUMENT, "Invalid Input Partition")
		logger.Error("Invalid Input Partition: ", zap.Any("Error", ex))
		res := dstk.ChangeRes{Ex: ex.Ex, Success: false}
		return &res, nil
	}
	s.store.Merge(c1, c2, n)
	res := dstk.ChangeRes{Success: false}
	return &res, nil
}

func (s *shardStoreUpdater) FindPartition(_ context.Context, in *dstk.Find_Req) (*dstk.Find_Res, error) {
	logger := zap.L()
	jobId := in.GetJobId()
	key := in.GetKey()
	if jobId == 0 || key == nil {
		ex := core.NewErr(dstk.Ex_INVALID_ARGUMENT, "Invalid job id or key")
		logger.Error("Invalid Input Partition: ", zap.Any("Error", ex))
		res := dstk.Find_Res{Ex: ex.Ex}
		return &res, nil
	}
	partition, err := s.store.Find(jobId, key)
	if err != nil {
		ex := core.NewErr(dstk.Ex_NOT_FOUND, err.Error())
		logger.Error("SplitPartition Error: ", zap.Any("Error", ex))
		res := dstk.Find_Res{Ex: ex.Ex}
		return &res, nil
	}
	res := dstk.Find_Res{Par: partition}
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

func generatePartitionId() int64 {
	return rand.Int63()
}

func getUrl() string {
	return "unassigned"
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
