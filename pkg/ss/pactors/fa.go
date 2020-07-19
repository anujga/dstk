package pactors

import (
	"fmt"
	"go.uber.org/zap"
)

type followingActor struct {
	*PartRange
}

func (fa *followingActor) become() error {
	fa.smState = Follower
	fa.logger.Info("became", zap.String("smstate", fa.smState.String()), zap.Int64("id", fa.Id()))
	for m := range fa.mailBox {
		fmt.Println(m)
	}
	fa.smState = Completed
	return nil
}

