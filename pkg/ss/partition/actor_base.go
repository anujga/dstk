package partition

import (
	"github.com/anujga/dstk/pkg/ss/common"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"sync/atomic"
)

type actorBase struct {
	db *sqlx.DB
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
	//ab.logger.Debug("setting current state", zap.Int64("part", ab.id), zap.String("state", state.String()))
	_, err := ab.db.Query("update partition set current_state=$1 where id=$2", state.String(), ab.id)
	if err != nil {
		ab.logger.Warn("failed to update current state", zap.Int64("part", ab.id), zap.String("state", state.String()))
	}
	ab.smState.Store(state)
}
