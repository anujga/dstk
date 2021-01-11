package partition

import (
	"github.com/anujga/dstk/pkg/core/control"
	"github.com/anujga/dstk/pkg/ss/common"
	"google.golang.org/grpc/status"
)

func handleReplicatedMsg(ab *actorBase, msg *common.ReplicatedMsg) {
	// todo no op for now
}

func handleClientMsg(ab *actorBase, cm common.ClientMsg) {
	//ab.logger.Info("client msg handling", zap.Int64("part", ab.id), zap.String("key", hex.EncodeToString(cm.Key())))
	res, err := ab.consumer.Process(cm)
	select {
	case cm.ResponseChannel() <- control.MaybeFailure(res, status.Convert(err)):
	default:
		// unlikely
	}
	close(cm.ResponseChannel())
}
