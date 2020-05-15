package sharder

import (
	"context"
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type shardLookupService struct {
	dstk.UnimplementedShardLookupServer
	clientShardStore *ClientShardStore
}

func (s *shardLookupService) FindPartition(_ context.Context, in *dstk.Find_Req) (*dstk.Find_Res, error) {
	logger := zap.L()
	jobId := in.GetJobId()
	key := in.GetKey()
	if jobId == 0 || key == nil {
		ex := core.NewErr(dstk.Ex_INVALID_ARGUMENT, "Invalid job id or key")
		logger.Error("Invalid job id or key: ", zap.Any("Error", ex))
		res := dstk.Find_Res{Ex: ex.Ex}
		return &res, nil
	}
	partition, err := s.clientShardStore.Find(jobId, key)
	if err != nil {
		ex := core.NewErr(dstk.Ex_NOT_FOUND, err.Error())
		logger.Error("FindPartition Error: ", zap.Any("Error", ex))
		res := dstk.Find_Res{Ex: ex.Ex}
		return &res, nil
	}
	res := dstk.Find_Res{Par: partition}
	return &res, nil
}

func createLookupService(port int32, jobs []int64) {
	logger := zap.L()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error("failed to listen: ", zap.Error(err))
	}
	s := grpc.NewServer()
	ssu := shardLookupService{
		clientShardStore: NewClientShardStore(jobs),
	}
	dstk.RegisterShardLookupServer(s, &ssu)
	logger.Info("Started Server ", zap.Any("Server:", s.GetServiceInfo()))
	logger.Info("Waiting for assignments")
	if err := s.Serve(lis); err != nil {
		logger.Error("failed to serve: ", zap.Error(err))
	}
}

func StartShardLookupService(port int32, jobs []int64) {
	initLogger()
	logger := zap.L()
	defer syncLogs(logger)
	// will return only when there is an error serving or graceful shutdown happens
	createLookupService(port, jobs)
	logger.Info("Shutting down the gRPC ShardLookupService")
}
