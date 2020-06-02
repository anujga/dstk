package main

import (
	"encoding/binary"
	badger "github.com/dgraph-io/badger/v2"
	"time"
)

type PersistentCounter struct {
	db *badger.DB
}

func (pc *PersistentCounter) Get(key string) (int64, error) {
	var res []byte
	err := pc.db.View(func(txn *badger.Txn) error {
		keyBytes := []byte(key)
		item, err := txn.Get(keyBytes)
		if err != nil {
			return err
		}
		res, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return 0, err
	}
	val, _ := binary.Varint(res)
	return val, nil
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
			existing, _ := item.ValueCopy(nil)
			exValue, _ = binary.Varint(existing)
		}
		res := make([]byte, 64)
		binary.PutVarint(res, value+exValue)
		entry := badger.NewEntry(keyBytes, res).WithTTL(time.Second * time.Duration(ttlSeconds))
		return txn.SetEntry(entry)
	})
}

func (pc *PersistentCounter) Remove(key string) error {
	return pc.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (pc *PersistentCounter) Close() error {
	return pc.db.Close()
}
