package bdb

import (
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/rangemap"
	"github.com/dgraph-io/badger/v2"
	"github.com/golang/protobuf/proto"
	"time"
)

type Wrapper struct {
	*badger.DB
}

// thread safe
func (w *Wrapper) Get(key []byte) (*dstk.DcDocument, error) {
	var res []byte
	var document *dstk.DcDocument
	err := w.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(key); err == nil {
			document = &dstk.DcDocument{}
			// TODO(gowri.sundaram): Remove value copy.
			res, err = item.ValueCopy(nil)
			err = proto.Unmarshal(res, document)
			return err
		} else {
			return err
		}
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, rangemap.ErrKeyAbsent(key).Err()
		}
		return nil, err
	}
	return document, nil
}

// thread safe
func (w *Wrapper) Put(key []byte, document *dstk.DcDocument, ttlSeconds float32) error {
	fmt.Printf("GOT REQUEST WITH ETAG!!!: %s\n", document.GetEtag())
	payload, err := proto.Marshal(document)
	if err != nil {
		return err
	}

	return w.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, payload).WithTTL(time.Duration(ttlSeconds) * time.Second)
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
