package bdb

import (
	"github.com/anujga/dstk/pkg/core"
	"github.com/dgraph-io/badger/v2"
	"google.golang.org/grpc/status"
	"time"
)

func lessOrEqHelper(txn *badger.Txn, k core.KeyT) ([]byte, error) {
	opts := badger.IteratorOptions{
		Reverse: true,
	}
	it := txn.NewIterator(opts)
	defer it.Close()

	it.Seek(k)
	if !it.Valid() {
		return nil, nil
	}

	return it.Item().ValueCopy(nil)
}

func LessOrEqual(db *badger.DB, name string, k core.KeyT) ([]byte, bool, *status.Status) {
	var res []byte

	err := db.View(func(txn *badger.Txn) error {
		r, err := lessOrEqHelper(txn, k)
		res = r
		return err
	})

	return res, res != nil, status.Convert(err)
}

type Wrapper struct {
	*badger.DB
}

// thread safe
func (w *Wrapper) Get(key []byte) ([]byte, error) {
	var res []byte
	err := w.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(key); err == nil {
			res, err = item.ValueCopy(nil)
			return err
		} else {
			return err
		}
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, core.ErrKeyAbsent(key).Err()
		}
		return nil, err
	}
	return res, nil
}

// thread safe
func (w *Wrapper) Put(key []byte, value []byte, ttlSeconds float32) error {
	return w.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, value).WithTTL(time.Duration(ttlSeconds) * time.Second)
		return txn.SetEntry(entry)
	})
}

// thread safe
func (w *Wrapper) Remove(key []byte) error {
	return w.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (w *Wrapper) StoreClose() error {
	return w.Close()
}
