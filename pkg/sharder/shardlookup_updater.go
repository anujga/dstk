package sharder

import (
	"context"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

type client struct {
	conn   *grpc.ClientConn
	c      *dstk.ShardStorageClient
	ctx    *context.Context
	cancel context.CancelFunc
}

func (store *ClientShardStore) initStore(jobId int64) {
	logger := zap.L()
	client := *store.ci.c
	req := dstk.Delta_Req{
		JobId:      jobId,
		FromTime:   0,
		ActiveOnly: true,
	}
	res, err := client.GetDeltaPartitions(*store.ci.ctx, &req)
	if err != nil {
		logger.Error("Init clientShardStore failed: ", zap.Error(err))
	} else {
		if res.GetEx() != nil {
			logger.Error("Init clientShardStore : Server responded with error: ",
				zap.Any(res.GetEx().GetMsg(), res.GetEx().GetId()))
		} else {
			if res.GetAdded() != nil {
				store.Update(jobId, res.GetAdded(), nil)
			}
		}
	}
	store.jobs = append(store.jobs, jobId)
}

func (store *ClientShardStore) StartUpdates(delay time.Duration) {
	ticker := time.NewTicker(delay)
	go func() {
		for {
			store.updateStore()
			select {
			case <-ticker.C:
			case <-store.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (store *ClientShardStore) StopCron() {
	store.stop <- true
}

func (store *ClientShardStore) updateStore() {
	logger := zap.L()
	client := *store.ci.c
	for _, jobId := range store.jobs {
		req := dstk.Delta_Req{
			JobId:      jobId,
			FromTime:   store.lastModified,
			ActiveOnly: true,
		}
		res, err := client.GetDeltaPartitions(*store.ci.ctx, &req)
		if err != nil {
			logger.Error("Update clientShardStore failed: ", zap.Error(err))
		} else {
			if res.GetEx() != nil {
				logger.Error("Update clientShardStore : Server responded with error: ",
					zap.Any(res.GetEx().GetMsg(), res.GetEx().GetId()))
			} else {
				if res.GetAdded() != nil || res.GetRemoved() != nil {
					store.Update(jobId, res.GetAdded(), res.GetRemoved())
				}
			}
		}
	}
}

func getClientInfo() *client {
	logger := zap.L()
	conn, err := grpc.Dial("addess", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.Error("Connect failed: ", zap.Error(err))
	}
	c := dstk.NewShardStorageClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	return &client{conn: conn, c: &c, ctx: &ctx, cancel: cancel}
}
