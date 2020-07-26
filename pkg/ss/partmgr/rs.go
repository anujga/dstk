package partition

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss/partition"
	"go.uber.org/zap"
)

var transitionTable = map[partition.State]map[partition.State]func(a partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) interface{}{
	partition.Init: {
		partition.CatchingUp: initToCatchingup,
		partition.Primary:    initToPrimary,
		partition.Follower:   initToFollower,
		partition.Proxy:      initToProxy,
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

type NewPart struct {
	partition.Actor
	*pb.Partition
}

func ensureActors(plist *pb.PartList, pm *managerImpl) []*NewPart {
	newActors := make([]*NewPart, 0)
	for _, part := range plist.GetParts() {
		if _, ok := pm.store.partIdMap[part.GetId()]; !ok {
			if c, maxOutstanding, err := pm.consumerFactory.Make(part); err == nil {
				pa := partition.NewActor(part, c, maxOutstanding)
				if e := pm.store.add(pa); e == nil {
					newActors = append(newActors, &NewPart{pa, part})
					pm.slog.Infow("part created", "id", part.GetId())
				} else {
					pm.slog.Errorw("failed to add part", "part", pa)
				}
			} else {
				pm.slog.Errorw("failed to make consumer", "part", part)
			}
		} else {
			pm.slog.Infow("partition already exists", "id", part.GetId())
		}
	}
	return newActors
}

func startNewActors(newParts []*NewPart, pm *managerImpl) {
	for _, newp := range newParts {
		currState := newp.Actor.State()
		desiredState := partition.StateFromString(newp.GetCurrentState())
		if currTo, ok := transitionTable[currState]; ok {
			if trf, ok := currTo[desiredState]; ok {
				msg := trf(newp.Actor, pm.store.partIdMap, newp.Partition)
				newp.Run(msg)
			} else {
				pm.slog.Warnw("no trans function", "from", currState.String(), "to", desiredState.String())
			}
		} else {
			pm.slog.Warnw("no trans function", "from", currState.String())
		}
	}
}

func resetParts(plist *pb.PartList, pm *managerImpl, logger *zap.Logger) error {
	newParts := ensureActors(plist, pm)
	startNewActors(newParts, pm)
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
	if currTo, ok := transitionTable[currState]; ok {
		if transFunc, ok := currTo[desiredState]; ok {
			logger.Info("state transition", zap.Int64("id", part.GetId()), zap.String("from", currState.String()), zap.String("to", desiredState.String()))
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