package partition

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss/partition"
	"github.com/anujga/dstk/pkg/ss/psm"
	"go.uber.org/zap"
)

type PartCombo struct {
	partition.Actor
	*pb.Partition
}

func ensureActors(plist *pb.Partitions, pm *managerImpl, partRpc pb.PartitionRpcClient) ([]*PartCombo, []*PartCombo) {
	newActors := make([]*PartCombo, 0)
	existingActors := make([]*PartCombo, 0)
	for _, part := range plist.GetParts() {
		if existingActor, ok := pm.store.partIdMap[part.GetId()]; !ok {
			if c, maxOutstanding, err := pm.consumerFactory.Make(part); err == nil {
				pa := partition.NewActor(part, c, maxOutstanding, partRpc)
				if e := pm.store.add(pa); e == nil {
					newActors = append(newActors, &PartCombo{pa, part})
					pm.slog.Infow("part created", "id", part.GetId())
				} else {
					pm.slog.Errorw("failed to add part", "part", pa)
				}
			} else {
				pm.slog.Errorw("failed to make consumer", "part", part)
			}
		} else {
			existingActors = append(existingActors, &PartCombo{
				Actor:     existingActor,
				Partition: part,
			})
			pm.slog.Debugw("partition already exists", "id", part.GetId())
		}
	}
	return newActors, existingActors
}

func startNewActors(newParts []*PartCombo, pm *managerImpl) {
	for _, newp := range newParts {
		currActorState := newp.Actor.State()
		currDbState := partition.StateFromString(newp.GetCurrentState())
		if currDbState == partition.Invalid {
			newp.Run(&partition.BecomeMsgImpl{TargetState: partition.Init})
		} else {
			if currTo, ok := psm.TransitionTable[currActorState]; ok {
				if trf, ok := currTo[currDbState]; ok {
					msg := trf(newp.Actor, pm.store.partIdMap, newp.Partition)
					newp.Run(msg)
				} else {
					pm.slog.Warn("no trans function", "from", currActorState.String(), "to", currDbState.String())
				}
			} else {
				pm.slog.Warnw("no trans function", "from", currActorState.String())
			}
		}
	}
}

func resetParts(plist *pb.Partitions, pm *managerImpl, logger *zap.Logger, partRpc pb.PartitionRpcClient) error {
	newParts, _ := ensureActors(plist, pm, partRpc)
	startNewActors(newParts, pm)
	for _, part := range plist.GetParts() {
		if currPa, ok := pm.store.partIdMap[part.GetId()]; ok {
			handleTransition(currPa, part, pm.store.partIdMap, logger)
		}
	}
	// todo handle partitions that are no longer present in new list
	return nil
}

func handleTransition(currPa partition.Actor, part *pb.Partition, pmap map[int64]partition.Actor, logger *zap.Logger) error {
	currState := currPa.State()
	desiredState := partition.StateFromString(part.GetDesiredState())
	if currState == desiredState {
		logger.Debug("state not changed", zap.Int64("part", part.GetId()), zap.Stringer("state", currState))
		return nil
	}
	if currTo, ok := psm.TransitionTable[currState]; ok {
		if transFunc, ok := currTo[desiredState]; ok {
			logger.Info("state transition", zap.Int64("id", part.GetId()), zap.Stringer("from", currState), zap.Stringer("to", desiredState))
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
			logger.Sugar().Warnw("no trans function", "from", currState.String(), "to", desiredState.String())
			// todo
		}
	} else {
		logger.Sugar().Warnw("no trans function", "from", currState.String())
		// todo
	}
	return nil
}
