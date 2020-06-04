package bdb

import (
	"github.com/dgraph-io/badger/v2"
	"time"
)

type Wrapper struct {
	*badger.DB
}

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
		return nil, err
	}
	return res, nil
}

func (w *Wrapper) Put(key []byte, value []byte, ttlSeconds float64) error {
	return w.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, value).WithTTL(time.Duration(ttlSeconds) * time.Second)
		return txn.SetEntry(entry)
	})
}

func (w *Wrapper) Remove(key []byte) error {
	return w.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (w *Wrapper) Close() error {
	return w.Close()
}
