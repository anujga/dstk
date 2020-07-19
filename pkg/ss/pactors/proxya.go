package pactors

import (
	"fmt"
	"go.uber.org/zap"
)

type proxyActor struct {
	*PartRange
}

func (pa *proxyActor) become() error {
	pa.smState = Proxy
	pa.logger.Info("became", zap.String("smstate", pa.smState.String()), zap.Int64("id", pa.Id()))
	for m := range pa.mailBox {
		fmt.Println(m)
	}
	pa.smState = Completed
	return nil
}
