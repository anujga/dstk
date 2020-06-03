package main

import (
	"encoding/binary"
	badger "github.com/dgraph-io/badger/v2"
	"time"
)

func counterMerge(old, incValue []byte) []byte {
	oldInt, _ := binary.Varint(old)
	incInt, _ := binary.Varint(incValue)
	res := make([]byte, 64)
	binary.PutVarint(res, oldInt+incInt)
	return res
}

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

func (pc *PersistentCounter) Inc(key string, value int64) error {
	incValBytes := make([]byte, 64)
	binary.PutVarint(incValBytes, value)
	mergeOp := pc.db.GetMergeOperator([]byte(key), counterMerge, time.Second*1)
	err := mergeOp.Add(incValBytes)
	mergeOp.Stop()
	return err
}

func (pc *PersistentCounter) Close() error {
	return pc.db.Close()
}
