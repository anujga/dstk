package se

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"go.uber.org/zap"
	"time"
)

type tc struct {
	notifications chan interface{}
	cache         stateHolder
	rpc           pb.SeClientApiClient
	lastModified  int64
	clientId      string
}

func (t *tc) Notifications() <-chan interface{} {
	return t.notifications
}

func (t *tc) Get(ctx context.Context, key []byte) (*pb.Partition, error) {
	return t.cache.Get(key)
}

func (t *tc) Parts() ([]*pb.Partition, error) {
	return t.cache.Parts()
}

func NewThickClient(clientId string, rpc pb.SeClientApiClient) ThickClient {
	t := tc{
		notifications: make(chan interface{}, 2),
		rpc:           rpc,
		clientId:      clientId,
	}

	core.Repeat(1*time.Minute, func(timestamp time.Time) bool {
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
	})
	t.cache.Clear()
	return &t
}

//todo: this should be a push instead of poll
func (t *tc) syncSe() error {
	rs, err := t.rpc.AllParts(context.TODO(), &pb.AllPartsReq{ClientId: t.clientId})
	if err != nil {
		return err
	}

	newTime := rs.GetLastModified()
	if newTime <= t.lastModified {
		return nil
	}
	err = t.cache.UpdateTree(rs.GetParts(), newTime)
	if err != nil {
		return err
	}

	return nil
}
