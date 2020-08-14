package bdb

import (
	"fmt"
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
	return document, err
}

// thread safe
func (w *Wrapper) Put(key []byte, document *dstk.DcDocument, ttlSeconds float32) error {
	fmt.Printf("Got document: %s\n", document)

	// We need to fetch current document to compare etag.
	currentDocument, err := w.Get(key)
	newDocument := (err == badger.ErrKeyNotFound)
	if err != nil && !newDocument {
		return err
	}
	if !newDocument && (document.GetEtag() != currentDocument.GetEtag()) {
		fmt.Printf("Expected etag: %s, received: %s\n", currentDocument.GetEtag(), document.GetEtag())
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
