package partition

import (
	"github.com/anujga/dstk/pkg/ss/common"
	"go.uber.org/zap"
	"sync/atomic"
)

type actorBase struct {
	id int64
	logger *zap.Logger
	smState *atomic.Value
	mailBox chan interface{}
	consumer common.Consumer
}

func (ab actorBase) getState() State {
	return ab.smState.Load().(State)
}

func (ab actorBase) setState(state State) {
	ab.smState.Store(state)
}
