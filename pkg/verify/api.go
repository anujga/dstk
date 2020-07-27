package verify

import (
	"context"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

type Process interface {
	Invoke(ctx context.Context) error
	Done(ctx context.Context) bool
	Init(ctx context.Context) error
}

//todo: replace with tracing/prometheus
type ProcessStats struct {
	Success, Failure int64
}

func RunProcess(p Process) *ProcessStats {
	ctx := context.TODO()
	s := ProcessStats{}
	if e := p.Init(ctx); e != nil {
		s.Failure += 1
		zap.S().Errorw("Failed to init", "err", e)
	}
	for !p.Done(ctx) {
		err := p.Invoke(ctx)
		if err != nil {
			s.Failure += 1
			zap.S().Errorw("invoke failed",
				"err", err)
			time.Sleep(1 * time.Second)
		} else {
			s.Success += 1
		}
	}
	return &s
}

type SampledProcess struct {
	Ps  []Process
	Rnd rand.Source
}

func (p *SampledProcess) Init(ctx context.Context) error {
	for _, p := range p.Ps {
		if e := p.Init(ctx); e != nil {
			return e
		}
	}
	return nil
}

func (p *SampledProcess) Invoke(ctx context.Context) error {
	n := int64(len(p.Ps))
	i := p.Rnd.Int63() % n
	proc := p.Ps[i]
	err := proc.Invoke(ctx)
	if err != nil {
		return err
	}

	if proc.Done(ctx) {
		if n == 1 {
			p.Ps = nil
		} else {
			p.Ps[i] = p.Ps[n-1]
			p.Ps = p.Ps[:n-1]
		}
	}
	return nil
}

func (p *SampledProcess) Done(context.Context) bool {
	return p.Ps == nil
}
