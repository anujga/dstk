package main

import (
	"encoding/binary"
	"fmt"
	badger "github.com/dgraph-io/badger/v2"
	"os"
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
		val, _ := binary.Varint(res)
		fmt.Printf("inside res %d\n", val)
		return err
	})
	if err != nil {
		return 0, err
	}
	val, _ := binary.Varint(res)
	return val, nil
}

func (pc *PersistentCounter) Inc(key string, value int64) error {
	mergeOp := pc.db.GetMergeOperator([]byte(key), counterMerge, time.Millisecond*1)
	defer mergeOp.Stop()
	incValBytes := make([]byte, 64)
	binary.PutVarint(incValBytes, value)
	err := mergeOp.Add(incValBytes)
	return err
}

func (pc *PersistentCounter) Close() error {
	return pc.db.Close()
}

func NewCounter(dbPath string) (*PersistentCounter, error) {
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, err
	}
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create db %s", err)
	}
	return &PersistentCounter{db: db}, nil
}

//func main() {
//	pc, err := NewCounter("/var/tmp/test-db")
//	fmt.Println(err)
//	defer pc.Close()
//	val, err := pc.Get("foo")
//	fmt.Println(val)
//	err = pc.Inc("foo", 10)
//	val, err = pc.Get("foo")
//	fmt.Println(val)
//	err = pc.Inc("foo", 10)
//	val, err = pc.Get("foo")
//	fmt.Println(val)
//}
