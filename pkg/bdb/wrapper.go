package bdb

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/dgraph-io/badger/v2"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"time"
)

type Wrapper struct {
	*badger.DB
}

// thread safe
func (w *Wrapper) Get(key []byte) (*dstk.DcDocument, error) {
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

	document := &dstk.DcDocument{}
	err = proto.Unmarshal(res, document)
	return document, err
}

// thread safe
func (w *Wrapper) Put(key []byte, document *dstk.DcDocument, ttlSeconds float32) error {
	return w.Update(func(txn *badger.Txn) error {
		// We need to fetch current document to compare etag.
		currentDocument, err := w.Get(key)
		newDocument := err == badger.ErrKeyNotFound
		if err != nil && !newDocument {
			return err
		}
		if !newDocument && (document.GetEtag() != currentDocument.GetEtag()) {
			return badger.ErrConflict
		}

		// Set a new etag.
		randomEtag, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		document.Etag = randomEtag.String()

		// Write to badger.
		payload, err := proto.Marshal(document)
		if err != nil {
			return err
		}
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