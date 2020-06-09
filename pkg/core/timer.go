package core

import (
	"time"
)

type Repeater struct {
	fn     func(time.Time) bool
	ticker *time.Ticker
	done   chan bool
}

func Repeat(freq time.Duration, fn func(time.Time) bool) *Repeater {
	ticker := time.NewTicker(freq)
	r := &Repeater{
		fn:     fn,
		ticker: ticker,
		done:   make(chan bool),
	}

	go func() {
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

func (r *Repeater) Stop() {
	r.ticker.Stop()
	r.done <- true
}
