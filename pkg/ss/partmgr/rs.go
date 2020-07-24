package partition

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss/partition"
	"go.uber.org/zap"
)

var transitionTable = map[partition.State]map[partition.State]func(a partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) interface{} {
	partition.Init: {
		partition.Follower: initToFollower,
		partition.Primary: initToPrimary,
	},
	partition.Primary: {
		partition.Proxy: primaryToProxy,
	},
	partition.Follower: {
		partition.Primary: followerToPrimary,
	},
	partition.Proxy: {
		partition.Retired: proxyToRetired,
	},
}

func ensureActors(plist *pb.PartList, pm *managerImpl)  {
	for _, part := range plist.GetParts() {
		if c, maxOutstanding, err := pm.consumerFactory.Make(part); err == nil {
			pa := partition.NewActor(part, c, maxOutstanding)
			if e := pm.store.add(pa); e == nil {
				pa.Run()
			} else {
				pm.slog.Errorw("failed to add part", "part", pa)
			}
		} else {
			pm.slog.Errorw("failed to make consumer", "part", part)
		}
	}
}

func resetParts(plist *pb.PartList, pm *managerImpl, logger *zap.Logger) error {
	ensureActors(plist, pm)
	for _, part := range plist.GetParts() {
		if currPa, ok := pm.store.partIdMap[part.GetId()]; ok {
			handleTransition(currPa, part, pm.store.partIdMap, logger)
		}
	}
	return nil
}

func handleTransition(currPa partition.Actor, part *pb.Partition, pmap map[int64]partition.Actor, logger *zap.Logger) error {
	currState := currPa.State()
	desiredState := partition.StateFromString(part.GetDesiredState())
	logger.Info("state transition", zap.Int64("id", part.GetId()), zap.String("from", currState.String()), zap.String("to", desiredState.String()))
	if currTo, ok := transitionTable[currState]; ok {
		if transFunc, ok := currTo[desiredState]; ok {
			if msg := transFunc(currPa, pmap, part); msg == nil {
				// todo
			} else {
				select {
				case currPa.Mailbox() <- msg:
				default:
					// todo
				}
			}
		} else {
			// todo
		}
	} else {
		// todo
	}
	return nil
}
