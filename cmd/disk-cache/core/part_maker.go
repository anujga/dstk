package dc

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/bdb"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/common"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"go.uber.org/zap"
	"os"
)

// 3. implement ss.ConsumerFactory
type partitionConsumerMaker struct {
	db             *bdb.Wrapper
	maxOutstanding int
}

func (m *partitionConsumerMaker) Make(p *dstk.Partition) (common.Consumer, int, error) {
	return &partitionConsumer{
		p:      p,
		pc:     m.db,
		logger: zap.L(),
		clock: &core.RealClock{},
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

func newConsumerMaker(dbPath string, maxOutstanding int) (*partitionConsumerMaker, error) {
	db, err := getDb(dbPath)
	if err != nil {
		return nil, err
	}
	return &partitionConsumerMaker{
		db:             &bdb.Wrapper{DB: db},
		maxOutstanding: maxOutstanding,
	}, nil
}
