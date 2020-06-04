package main

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/bdb"
	"github.com/anujga/dstk/pkg/ss"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"os"
)

// 3. implement ss.ConsumerFactory
type partitionCounterMaker struct {
	db             *bdb.Wrapper
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
		db:             &bdb.Wrapper{DB: db},
		maxOutstanding: maxOutstanding,
	}, nil
}
