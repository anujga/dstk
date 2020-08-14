package split

import (
	"context"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/partition"
	"go.uber.org/zap"
	"time"
)

type Dag struct {
	Part        *dstk.Partition
	IdGenerator core.IdGenerator
	SeRpc       dstk.PartitionRpcClient
	slog        *zap.SugaredLogger
}

func (d *Dag) Start(ctx context.Context, splitPoint core.KeyT) error {
	d.slog.Infow("split started", "part", d.Part, "split", splitPoint)
	partId := d.Part.GetId()
	split1Id := d.IdGenerator()
	d.createFollower(split1Id, partId, d.Part.GetStart(), splitPoint)

	split2Id := d.IdGenerator()
	d.createFollower(split2Id, partId, splitPoint, d.Part.GetEnd())

	// todo handle errors
	d.checkState(split1Id, partition.Follower, ctx)
	d.checkState(split2Id, partition.Follower, ctx)

	d.makeProxy(partId, []int64{split1Id, split2Id})
	d.checkState(partId, partition.Proxy, ctx)

	d.make(split1Id, partition.Primary)
	d.make(split2Id, partition.Primary)
	d.checkState(split1Id, partition.Primary, ctx)
	d.checkState(split2Id, partition.Primary, ctx)

	d.make(partId, partition.Retired)
	d.checkState(partId, partition.Retired, ctx)
	d.slog.Infow("split ended", "part id", d.Part.GetId())
	return nil
}

func (d *Dag) checkState(partId int64, state partition.State, ctx context.Context) error {
	t := time.NewTicker(time.Second * 5)
	for {
		select {
		case _ = <-t.C:
			res, err := d.SeRpc.GetPartitions(ctx, &dstk.PartitionGetRequest{
				Id: partId,
			})
			if err != nil {
				d.slog.Errorw("getting partitions failed", "part id", partId)
				return err
			}
			if res.GetPartitions().GetParts()[0].GetDesiredState() == state.String() {
				return nil
			}
		}
	}
}

func (d *Dag) createFollower(partId, leaderId int64, start, end []byte) {

}

func (d *Dag) makeProxy(partId int64, proxyTo []int64) {

}

func (d *Dag) make(partId int64, state partition.State) {

}
