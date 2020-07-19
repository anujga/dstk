package partmgr

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/common"
	"github.com/anujga/dstk/pkg/ss/partition"
	"github.com/google/btree"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

type PartManager interface {
	Find(key core.KeyT) (partition.Actor, error)
	Reset(plist *pb.PartList) error
}

// todo: this is the 4th implementation of the range map.
// need to define a proper data structure that can be reused
type PartManagerImpl struct {
	store           *partRangeStore
	consumerFactory common.ConsumerFactory
	slog            *zap.SugaredLogger
}

func (pm *PartManagerImpl) Find(key core.KeyT) (partition.Actor, error) {
	return pm.store.find(key)
}

func (pm *PartManagerImpl) Reset(plist *pb.PartList) error {
	for _, part := range plist.GetParts() {
		if currPa, ok := pm.store.partIdMap[part.GetId()]; ok {
			// todo these are not handled by various partition actors. this notifies the existing actor to
			// become the actor in the partition sent
			currPa.Mailbox() <- part
		} else {
			if c, maxOutstanding, err := pm.consumerFactory.Make(part); err == nil {
				var leader partition.Actor
				if part.GetLeaderId() != 0 {
					leader = pm.store.partIdMap[part.GetLeaderId()]
				}
				pa := partition.NewPartActor(part, c, maxOutstanding, leader)
				pa.Run()
				if e := pm.store.add(pa); e != nil {
					pm.slog.Errorw("failed to add part", "part", pa)
				}
			} else {
				pm.slog.Errorw("failed to make consumer", "part", part)
			}
		}
	}
	return nil
}

//todo: ensure there is at least 1 partition during construction
func NewPartitionMgr(factory common.ConsumerFactory) (PartManager, *status.Status) {
	return &PartManagerImpl{
		consumerFactory: factory,
		slog:            zap.S(),
		store: &partRangeStore{
			partRoot:     btree.New(16),
			partIdMap:    make(map[int64]partition.Actor),
			lastModified: 0,
		},
	}, nil
}
