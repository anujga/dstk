package partition

import (
	"github.com/anujga/dstk/pkg/ss/common"
)

func handleReplicatedMsg(ab *actorBase, rm *common.ReplicatedMsg) {
	// todo no op for now
}

func handleClientMsg(ab *actorBase, cm common.ClientMsg) {
	//ab.logger.Debug("client msg handling", zap.Int64("part", ab.id), zap.String("key", hex.EncodeToString(cm.Key())))
	res, err := ab.consumer.Process(cm)
	select {
	case cm.ResponseChannel() <- &common.Response{
		Res: res,
		Err: err,
	}:
	default:
		// unlikely
	}
	close(cm.ResponseChannel())
}
