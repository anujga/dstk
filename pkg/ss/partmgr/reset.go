package partitionmgr

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss/partition"
	"github.com/anujga/dstk/pkg/ss/psm"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

type PartCombo struct {
	partition.Actor
	*pb.Partition
}

func ensureActors(plist *pb.Partitions, pm *managerImpl, partRpc pb.PartitionRpcClient) ([]*PartCombo, []*PartCombo) {
	newActors := make([]*PartCombo, 0)
	slog := zap.S()
	existingActors := make([]*PartCombo, 0)
	for _, part := range plist.GetParts() {
		existingActor, found := pm.store.partIdMap[part.GetId()]
		if !found {
			if c, maxOutstanding, err := pm.consumerFactory.Make(part); err == nil {
				pa := partition.NewActor(part, c, maxOutstanding, partRpc)
				if e := pm.store.add(pa); e == nil {
					newActors = append(newActors, &PartCombo{pa, part})
					slog.Infow("part created", "id", part.GetId())
				} else {
					slog.Errorw("failed to add part", "part", pa)
				}
			} else {
				slog.Errorw("failed to make consumer", "part", part)
			}
		} else {
			existingActors = append(existingActors, &PartCombo{
				Actor:     existingActor,
				Partition: part,
			})
			slog.Debugw("partition already exists", "id", part.GetId())
		}
	}
	return newActors, existingActors
}

func startNewActors(newParts []*PartCombo, pm *managerImpl) {
	slog := zap.S()

	for _, newp := range newParts {
		currActorState := newp.Actor.State()
		currDbState := partition.StateFromString(newp.GetCurrentState())
		// todo: Not sure invalid is same as init
		if currDbState == partition.Invalid {
			newp.Run(&partition.BecomeMsgImpl{TargetState: partition.Init})
			continue

		}
		trf, st := psm.GetTransition(currActorState, currDbState)
		if st != nil {
			//todo: this is a fatal error.
			slog.Warnw("no trans function",
				"err", st)
			continue
		}
		msg := trf(newp.Actor, pm.store.partIdMap, newp.Partition)
		newp.Run(msg)
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

func handleTransition(currPa partition.Actor, part *pb.Partition, pmap map[int64]partition.Actor, logger *zap.Logger) *status.Status {
	currState := currPa.State()
	desiredState := partition.StateFromString(part.GetDesiredState())
	if currState == desiredState {
		logger.Debug("state not changed",
			zap.Int64("part", part.GetId()),
			zap.Stringer("state", currState))
		return nil
	}
	slog := zap.S()

	transFunc, st := psm.GetTransition(currState, desiredState)
	if st != nil {
		//todo: this is a fatal error.
		slog.Warnw("no trans function",
			"err", st)
		return st
	}

	logger.Info("state transition",
		zap.Int64("id", part.GetId()),
		zap.Stringer("from", currState),
		zap.Stringer("to", desiredState))

	msg := transFunc(currPa, pmap, part)

	if msg == nil {
		// todo
	} else {
		select {
		case currPa.Mailbox() <- msg:
		default:
			// todo
		}
	}

	return nil
}
