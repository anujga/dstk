package se

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type tc struct {
	notifications chan interface{}
	cache         stateHolder
	rpc           pb.PartitionRpcClient
	clientId      string
}

func (t *tc) Notifications() <-chan interface{} {
	return t.notifications
}

func (t *tc) Get(ctx context.Context, key []byte) (*pb.Partition, *status.Status) {
	return t.cache.Get(key)
}

func (t *tc) Parts() ([]*pb.Partition, error) {
	return t.cache.Parts()
}

func NewThickClient(clientId string, rpc pb.PartitionRpcClient) (ThickClient, *status.Status) {
	t := tc{
		notifications: make(chan interface{}, 2),
		rpc:           rpc,
		clientId:      clientId,
	}

	rep := core.Repeat(5*time.Second, func(timestamp time.Time) bool {
		err := t.syncSe()
		if err != nil {
			zap.S().Errorw("fetch updates from SE",
				"err", err)
		} else {
			delay := timestamp.UnixNano() - t.cache.LastModified()
			zap.S().Infow("fetch updates from SE",
				"time", timestamp,
				"delay", delay)
		}
		return true
	}, true)

	if rep == nil {
		return nil, core.ErrInfo(
			codes.Internal,
			"failed to initialize via se",
			"se", rpc)
	}

	return &t, nil
}

//todo: this should be a push instead of poll
func (t *tc) syncSe() error {
	rs, err := t.rpc.GetPartitions(context.TODO(), &pb.PartitionGetRequest{FetchAll: true})
	if err != nil {
		return err
	}

	newTime := int64(0)
	for _, p := range rs.Partitions.GetParts() {
		if p.GetModifiedOn() > newTime {
			newTime = p.GetModifiedOn()
		}
	}
	if newTime <= t.cache.LastModified() {
		return nil
	}
	err = t.cache.UpdateTree(rs.GetPartitions(), newTime)
	if err != nil {
		return err
	}

	return nil
}
