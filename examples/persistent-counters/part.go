package main

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss"
	badger "github.com/dgraph-io/badger/v2"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"time"
)

var dbLatencySummary = prometheus.NewSummary(prometheus.SummaryOpts{
	Name:       "db_latency_seconds",
	Objectives: map[float64]float64{0.9: 0.01, 0.99: 0.001},
})

// 2. Define the state for a given partition and implement ss.Consumer
type partitionCounter struct {
	p  *dstk.Partition
	pc *PersistentCounter
}

func (m *partitionCounter) Meta() *dstk.Partition {
	return m.p
}

/// this method does not have to be thread safe
func (m *partitionCounter) Process(msg0 ss.Msg) bool {
	//go func() {
	msg := msg0.(*Request)
	var err error
	c := msg.ResponseChannel()
	// TODO better way to model get/inc requests
	if msg.V == 0 {
		if val, err := m.pc.Get(msg.K); err == nil {
			c <- val
		} else {
			c <- err
		}
	} else {
		err = func() error {
			t := time.Now()
			defer func() {
				dbLatencySummary.Observe(time.Since(t).Seconds())
			}()
			//t := prometheus.NewTimer(dbLatencyHistogram)
			//defer t.ObserveDuration()
			//t := time.Now()
			e := m.pc.Inc(msg.K, msg.V)
			//fmt.Println(time.Now().Sub(t))
			return e
		}()
		if err == nil {
			c <- "counter incremented"
		} else {
			c <- err
		}
	}
	close(c)
	//}()
	return true
}

// 3. implement ss.ConsumerFactory

type partitionCounterMaker struct {
	db             *badger.DB
	maxOutstanding int
}

func (m *partitionCounterMaker) Make(p *dstk.Partition) (ss.Consumer, int, error) {
	pc := &PersistentCounter{db: m.db}
	return &partitionCounter{
		p:  p,
		pc: pc,
	}, m.maxOutstanding, nil
}

func getDb(dbPath string) (*badger.DB, error) {
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, err
	}
	// TODO: gracefully stop the db too
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		return nil, err
	}
	return db, err
}

func newCounterMaker(dbPath string, maxOutstanding int) (*partitionCounterMaker, error) {
	db, err := getDb(dbPath)
	if err != nil {
		return nil, err
	}
	return &partitionCounterMaker{
		db:             db,
		maxOutstanding: maxOutstanding,
	}, nil
}
