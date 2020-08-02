package partition

import (
	"encoding/hex"
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
)

func handleReplicatedMsg(ab *actorBase, rm *common.ReplicatedMsg)  {
	// todo no op for now
}

func handleClientMsg(ab *actorBase, cm common.ClientMsg)  {
	ab.logger.Debug("client msg handling", zap.Int64("part", ab.id), zap.String("key", hex.EncodeToString(cm.Key())))
	res, err := ab.consumer.Process(cm)
	resC := cm.ResponseChannel()
	if err != nil {
		resC <- err
	} else {
		resC <- res
	}
	close(resC)
}
