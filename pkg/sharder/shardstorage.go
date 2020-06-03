package sharder

import (
	"context"
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
)

type shardStoreUpdater struct {
	dstk.UnimplementedShardStorageServer
	store *ShardStore
}

func (s *shardStoreUpdater) CreatePartition(_ context.Context, in *dstk.CreateJobPartReq) (*dstk.ChangeRes, error) {
	logger := zap.L()
	jobId := in.GetJobId()
	markings := in.GetMarkings()
	if jobId == 0 || markings == nil || len(markings) < 1 {
		ex := core.NewErr(dstk.Ex_INVALID_ARGUMENT, "Invalid markings")
		logger.Error("Invalid markings: ", zap.Any("Error", ex))
		res := dstk.ChangeRes{Ex: ex.Ex, Success: false}
		return &res, nil
	}
	err := s.store.Create(jobId, markings)
	if err != nil {
		ex := core.NewErr(dstk.Ex_CONFLICT, err.Error())
		logger.Error("CreatePartition Error: ", zap.Any("Error", ex))
		res := dstk.ChangeRes{Ex: ex.Ex, Success: false}
		return &res, nil
	}
	res := dstk.ChangeRes{Success: true}
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
	jobId := in.GetJobId()
	c1 := in.GetC1()
	c2 := in.GetC2()
	if jobId == 0 || c1 == nil || c2 == nil {
		ex := core.NewErr(dstk.Ex_INVALID_ARGUMENT, "Invalid Input Partition")
		logger.Error("Invalid Input Partition: ", zap.Any("Error", ex))
		res := dstk.ChangeRes{Ex: ex.Ex, Success: false}
		return &res, nil
	}
	err := s.store.Merge(jobId, c1, c2)
	if err != nil {
		ex := core.NewErr(dstk.Ex_BAD_PARTITION, err.Error())
		logger.Error("MergePartition Error: ", zap.Any("Error", ex))
		res := dstk.ChangeRes{Ex: ex.Ex, Success: false}
		return &res, nil
	}
	res := dstk.ChangeRes{Success: true}
	return &res, nil
}

//
//func (s *shardStoreUpdater) FindPartition(_ context.Context, in *dstk.Find_Req) (*dstk.Find_Res, error) {
//	logger := zap.L()
//	jobId := in.GetJobId()
//	key := in.GetKey()
//	if jobId == 0 || key == nil {
//		ex := core.NewErr(dstk.Ex_INVALID_ARGUMENT, "Invalid job id or key")
//		logger.Error("Invalid Input Partition: ", zap.Any("Error", ex))
//		res := dstk.Find_Res{Ex: ex.Ex}
//		return &res, nil
//	}
//	partition, err := s.store.Find(jobId, key)
//	if err != nil {
//		ex := core.NewErr(dstk.Ex_NOT_FOUND, err.Error())
//		logger.Error("SplitPartition Error: ", zap.Any("Error", ex))
//		res := dstk.Find_Res{Ex: ex.Ex}
//		return &res, nil
//	}
//	res := dstk.Find_Res{Par: partition}
//	return &res, nil
//}

func (s *shardStoreUpdater) GetDeltaPartitions(_ context.Context, in *dstk.Delta_Req) (*dstk.Delta_Res, error) {
	logger := zap.L()
	jobId := in.GetJobId()
	if jobId == 0 {
		ex := core.NewErr(dstk.Ex_INVALID_ARGUMENT, "Invalid job id or key")
		logger.Error("Invalid argument: ", zap.Any("Error", ex))
		res := dstk.Delta_Res{Ex: ex.Ex}
		return &res, nil
	}
	partitions, err := s.store.GetDelta(in.GetJobId(), in.GetFromTime(), in.GetActiveOnly())
	if err != nil {
		ex := core.NewErr(dstk.Ex_NOT_FOUND, err.Error())
		logger.Error("GetDeltaPartitions Error: ", zap.Any("Error", ex))
		res := dstk.Delta_Res{Ex: ex.Ex}
		return &res, nil
	}
	added := ([]*dstk.Partition)(nil)
	removed := ([]*dstk.Partition)(nil)
	for _, part := range partitions {
		if part.GetActive() {
			added = append(added, part)
		} else {
			removed = append(removed, part)
		}
	}
	res := dstk.Delta_Res{
		Added:   added,
		Removed: removed,
	}
	return &res, nil
}

func initLogger() {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
}

func createServer(port int32) {
	logger := zap.L()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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

func syncLogs(logger *zap.Logger) {
	err := logger.Sync()
	if err != nil {
		log.Fatal("Error syncing log!", err)
	}
}

func StartShardStorageServer(port int32) {
	initLogger()
	logger := zap.L()
	defer syncLogs(logger)
	// will return only when there is an error serving or graceful shutdown happens
	createServer(port)
	logger.Info("Shutting down the gRPC shardStoreUpdater")
}
