package rangemap

import (
	"github.com/anujga/dstk/pkg/core"
	"github.com/dgraph-io/badger/v2"
)

type badgerMap struct {
	db *badger.DB
}

func (b *badgerMap) Get(key core.KeyT) (Range, error) {
	panic("implement me")
}

func (b *badgerMap) Put(rng Range) error {
	panic("implement me")
}

func (b *badgerMap) Remove(rng Range) (Range, error) {
	panic("implement me")
}
