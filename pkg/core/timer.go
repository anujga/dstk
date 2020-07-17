package core

import (
	"go.uber.org/zap"
	"time"
)

type Repeater struct {
	fn     func(time.Time) bool
	ticker *time.Ticker
	done   chan bool
}

func Repeat(freq time.Duration, fn func(time.Time) bool, firstRunSync bool) *Repeater {
	ticker := time.NewTicker(freq)
	r := &Repeater{
		fn:     fn,
		ticker: ticker,
		done:   make(chan bool),
	}

	if firstRunSync {
		if !r.fn(time.Now()) {
			CloseLogErr(r)
			zap.S().Warnw("Repeater breaking on first run",
				"fn", fn)
			return nil
		}
	}

	go func() {
		defer CloseLogErr(r)

		for {
			select {
			case <-r.done:
				return
			case t := <-ticker.C:
				if !r.fn(t) {
					return
				}
			}
		}
	}()

	return r
}

func (r *Repeater) Close() error {
	r.ticker.Stop()
	r.done <- true
	return nil
}
