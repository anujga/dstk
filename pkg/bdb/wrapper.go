package bdb

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/dgraph-io/badger/v2"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/status"
	"time"
)

func iterHelper(txn *badger.Txn, reverse bool, k core.KeyT) ([]byte, error) {
	opts := badger.IteratorOptions{
		Reverse: reverse,
	}

	if len(k) == 0 {
		k = core.MinKey
	}

	it := txn.NewIterator(opts)
	defer it.Close()

	it.Seek(k)
	if !it.Valid() {
		return nil, nil
	}

	return it.Item().ValueCopy(nil)
}

// returns (eof reached, error) incase of error, eof flag is irrelevant
func Iter(txn *badger.Txn, startInclusive core.KeyT, reverse bool, fn func(item *badger.Item) (bool, error)) (bool, error) {
	//todo: get rid of min key check. modify test case and valid range
	if len(startInclusive) == 0 {
		startInclusive = core.MinKey
	}

	it := txn.NewIterator(badger.IteratorOptions{Reverse: reverse})
	defer it.Close()

	for it.Seek(startInclusive); it.Valid(); it.Next() {
		continueLoop, err := fn(it.Item())
		if err != nil {
			return false, err
		}
		if !continueLoop {
			return false, nil
		}
	}

	return true, nil
}

func IterNVals(txn *badger.Txn, startInclusive core.KeyT, reverse bool, n int) (bool, [][]byte, error) {
	var rs [][]byte
	eof, err := Iter(txn, startInclusive, reverse, func(item *badger.Item) (bool, error) {
		v, err := item.ValueCopy(nil)
		if err != nil {
			return false, err
		}
		rs = append(rs, v)
		n -= 1
		cont := n > 0
		return cont, nil
	})

	return eof, rs, err
}

func LessOrEqual(db *badger.DB, k core.KeyT) ([]byte, bool, *status.Status) {
	return FindFirst(db, true, k)
}

func EqualOrGreater(db *badger.DB, k core.KeyT) ([]byte, bool, *status.Status) {
	return FindFirst(db, false, k)
}

func FindFirst(db *badger.DB, reverse bool, k core.KeyT) ([]byte, bool, *status.Status) {
	var res []byte

	err := db.View(func(txn *badger.Txn) error {
		r, err := iterHelper(txn, reverse, k)
		res = r
		return err
	})

	return res, res != nil, status.Convert(err)
}

type Wrapper struct {
	*badger.DB
	core.EtagGenerator
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
		if !newDocument && (document.GetMeta().GetEtag() != currentDocument.GetMeta().GetEtag()) {
			return badger.ErrConflict
		}

		// Set a new etag.
		var newEtag int64
		if newDocument {
			newEtag = w.EtagGenerator.Initial()
		} else {
			newEtag = w.EtagGenerator.Next(currentDocument.GetMeta().GetEtag())
		}
		document.GetMeta().Etag = newEtag

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