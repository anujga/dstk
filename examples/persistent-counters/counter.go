package main

import (
	"encoding/binary"
	"fmt"
	"github.com/anujga/dstk/pkg/bdb"
	badger "github.com/dgraph-io/badger/v2"
	"time"
)

type PersistentCounter struct {
	db *bdb.Wrapper
}

func (pc *PersistentCounter) Get(key string) (int64, error) {
	if res, err := pc.db.Get([]byte(key)); err == nil {
		val, _ := binary.Varint(res)
		return val, nil
	} else {
		return 0, err
	}
}

func (pc *PersistentCounter) Inc(key string, value int64, ttlSeconds float64) error {
	incValBytes := make([]byte, 64)
	binary.PutVarint(incValBytes, value)
	keyBytes := []byte(key)
	return pc.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(keyBytes)
		exValue := int64(0)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == nil {
			if existingBytes, err := item.ValueCopy(nil); err != nil {
				return err
			} else {
				var n int
				if exValue, n = binary.Varint(existingBytes); n <= 0 {
					return fmt.Errorf("invalid data for %s", key)
				}
			}
		}
		res := make([]byte, 64)
		binary.PutVarint(res, value+exValue)
		entry := badger.NewEntry(keyBytes, res).WithTTL(time.Second * time.Duration(ttlSeconds))
		return txn.SetEntry(entry)
	})
}

func (pc *PersistentCounter) Remove(key string) error {
	return pc.db.Remove([]byte(key))
}

func (pc *PersistentCounter) Close() error {
	return pc.db.Close()
}
