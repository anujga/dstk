package main

import (
	"encoding/binary"
	"fmt"
	badger "github.com/dgraph-io/badger/v2"
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

func (pc *PersistentCounter) Inc(key string, value int64) error {
	presentVal, err := pc.Get(key)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			presentVal = 0
		} else {
			return err
		}
	}
	err = pc.db.Update(func(txn *badger.Txn) error {
		keyBytes := []byte(key)
		valBytes := make([]byte, 64)
		binary.PutVarint(valBytes, presentVal+value)
		err = txn.Set(keyBytes, valBytes)
		return err
	})
	return err
}

func (pc *PersistentCounter) Close() error {
	return pc.db.Close()
}

func NewCounter(dbPath string) (*PersistentCounter, error) {
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create db %s", err)
	}
	return &PersistentCounter{db: db}, nil
}

//
//func main() {
//	pc, err := NewCounter()
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
