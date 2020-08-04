package split

import (
	"context"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/partition"
)

type Dag struct {
	Part *dstk.Partition
	IdGenerator func() int64
	SeRpc dstk.PartitionRpcClient
}

func (d *Dag) Start(ctx context.Context, splitPoint []byte) *core.FutureErr {
	p := core.NewPromise()
	return p.Complete(func() error {
		partId := d.Part.GetId()
		split1Id := d.IdGenerator()
		d.createFollower(split1Id, partId, d.Part.GetStart(), splitPoint)

		split2Id := d.IdGenerator()
		d.createFollower(split2Id, partId, splitPoint, d.Part.GetEnd())

		d.checkState(split1Id, partition.Follower).Wait()
		d.checkState(split2Id, partition.Follower).Wait()

		d.makeProxy(partId, []int64{split1Id, split2Id})
		d.checkState(partId, partition.Proxy).Wait()

		d.make(split1Id, partition.Primary)
		d.make(split2Id, partition.Primary)
		d.checkState(split1Id, partition.Primary).Wait()
		d.checkState(split2Id, partition.Primary).Wait()

		d.make(partId, partition.Retired)
		d.checkState(partId, partition.Retired).Wait()
		return nil
	})
}

func (d *Dag) checkState(partId int64, state partition.State) *core.FutureErr {
	return nil
}

func (d *Dag) createFollower(partId, leaderId int64, start, end []byte) {

}

func (d *Dag) makeProxy(partId int64, proxyTo []int64)  {

}

func (d *Dag) make(partId int64, state partition.State)  {

}
