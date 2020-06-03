package main

import (
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss"
	badger "github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"time"
)

var dbLatencySummary = prometheus.NewSummary(prometheus.SummaryOpts{
	Name:       "db_latency_seconds",
	Objectives: map[float64]float64{0.9: 0.01, 0.99: 0.001},
})

var secondsInDay = (time.Second * 86400).Seconds()

// 2. Define the state for a given partition and implement ss.Consumer
type partitionCounter struct {
	p  *dstk.Partition
	pc *PersistentCounter
}

func (m *partitionCounter) Meta() *dstk.Partition {
	return m.p
}

func (m *partitionCounter) get(req *Request) bool {
	if val, err := m.pc.Get(req.K); err == nil {
		req.C <- val
		return true
	} else {
		req.C <- err
		return false
	}
}

func (m *partitionCounter) remove(req *Request) bool {
	if err := m.pc.Remove(req.K); err == nil {
		req.C <- fmt.Sprintf("%s removed", req.K)
		return true
	} else {
		req.C <- err
		return false
	}
}

func (m *partitionCounter) inc(req *Request) bool {
	t := time.Now()
	defer func() {
		dbLatencySummary.Observe(time.Since(t).Seconds())
	}()
	ttl := req.TtlSeconds
	if ttl == 0 {
		ttl = secondsInDay
	}
	e := m.pc.Inc(req.K, req.V, ttl)
	if e == nil {
		req.C <- fmt.Sprintf("%s incremented", req.K)
	} else {
		req.C <- e
	}
	return e == nil
}

/// this method does not have to be thread safe
func (m *partitionCounter) Process(msg0 ss.Msg) bool {
	msg := msg0.(*Request)
	c := msg.ResponseChannel()
	defer close(c)
	switch msg.RequestType {
	case Get:
		return m.get(msg)
	case Inc:
		return m.inc(msg)
	case Remove:
		return m.remove(msg)
	}
	return true
}

// 3. implement ss.ConsumerFactory

type partitionCounterMaker struct {
	db             *badger.DB
	maxOutstanding int
}

func (m *partitionCounterMaker) Make(p *dstk.Partition) (ss.PartHandler, int, error) {
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
	opt := badger.DefaultOptions(dbPath).
		WithTableLoadingMode(options.LoadToRAM).
		WithValueLogLoadingMode(options.MemoryMap)
	db, err := badger.Open(opt)
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
