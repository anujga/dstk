package node

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type PartsSyncer struct {
	wa    Actor
	seRpc pb.PartitionRpcClient
}

func (ps *PartsSyncer) Start() *status.Status {
	rep := core.Repeat(1*time.Minute, func(timestamp time.Time) bool {
		if err := ps.syncFromSe(); err == nil {
			zap.S().Infow("fetch updates from SE",
				"time", timestamp)
		} else {
			zap.S().Errorw("fetch updates from SE", "err", err)
		}
		return true
	}, true)
	if rep == nil {
		return core.ErrInfo(
			codes.Internal,
			"failed to initialize via se",
			"se", ps.seRpc)
	}
	return nil
}

func (ps *PartsSyncer) syncFromSe() error {
	req := &pb.PartitionGetRequest{
		WorkerId: int64(ps.wa.Id()),
	}
	newParts, err := ps.seRpc.GetPartitions(context.TODO(), req)
	if err != nil {
		return err
	}
	ps.wa.Mailbox() <- newParts.GetPartitions()
	return nil
}

func NewSyncer(wa Actor, seRpc pb.PartitionRpcClient) *PartsSyncer {
	return &PartsSyncer{
		wa:    wa,
		seRpc: seRpc,
	}
}
