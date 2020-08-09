package partition

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
	"sync/atomic"
)

type actorBase struct {
	id             int64
	logger         *zap.Logger
	smState        *atomic.Value
	currentDbState State
	mailBox        chan interface{}
	consumer       common.Consumer
	partitionRpc   pb.PartitionRpcClient
}

func (ab actorBase) getState() State {
	return ab.smState.Load().(State)
}

func (ab actorBase) setState(state State) error {
	if ab.currentDbState != state {
		ab.logger.Info("setting current state", zap.Int64("part", ab.id), zap.Stringer("state", state))
		req := &pb.PartitionUpdateRequest{
			Id:           ab.id,
			CurrentState: state.String(),
		}
		_, err := ab.partitionRpc.UpdatePartition(context.TODO(), req)
		if err != nil {
			ab.logger.Error("failed to set state", zap.Any("req", req))
			return err
		}
		ab.currentDbState = state
	}
	ab.smState.Store(state)
	return nil
}
