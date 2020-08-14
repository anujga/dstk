package rangemap

import (
	"github.com/anujga/dstk/pkg/bdb"
	"github.com/anujga/dstk/pkg/core"
	"github.com/dgraph-io/badger/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

type badgerMap struct {
	db      *badger.DB
	name    string
	marshal RangeEncoder
	mu      sync.Mutex
}

func (b *badgerMap) Close() error {
	return b.db.Close()
}

func (b *badgerMap) Get(key core.KeyT) (Range, bool, *status.Status) {
	if !core.ValidKey(key) {
		return nil, false, core.ErrInfo(
			codes.InvalidArgument,
			"invalid key",
			"key", key)

	}
	r, rangeFound, err := b.precedingRange(key)
	if err != nil {
		return nil, false, err
	}

	present := rangeFound && RangeContains(r, key)

	if !present {
		return nil, false, nil
	}

	return r, true, nil
}

func (b *badgerMap) precedingRange(key core.KeyT) (Range, bool, *status.Status) {
	return b.iterRange(key, true)
}

func (b *badgerMap) iterRange(key core.KeyT, reverse bool) (Range, bool, *status.Status) {
	item, found, err := bdb.FindFirst(b.db, reverse, key)
	if err != nil {
		return nil, false, err
	}

	if !found {
		return nil, false, nil
	}

	r, stat := b.marshal.Unmarshal(item)
	if stat != nil {
		return nil, false, status.Convert(stat)
	}

	return r, true, nil
}

func (b *badgerMap) Put(rng Range) *status.Status {
	if !ValidRange(rng) {
		return ErrInvalidRange(rng)
	}

	// todo: instead of taking a long transaction block,
	// we are only taking a lock here. replace with transaction
	b.mu.Lock()
	defer b.mu.Unlock()

	before, beforeFound, st := b.precedingRange(rng.Start())
	if st != nil {
		return st
	}

	if beforeFound {
		if RangeContains(before, rng.Start()) {
			return core.ErrInfo(codes.InvalidArgument,
				"New range overlaps",
				"existing", before,
				"new", rng)
		}
	}

	after, afterFound, st := b.iterRange(rng.Start(), false)

	if st != nil {
		return st
	}

	if afterFound {
		//if RangeContainsExcludeStart(after, rng.End()) {
		//	return core.ErrInfo(codes.InvalidArgument,
		//		"New range overlaps",
		//		"existing", after,
		//		"new", rng)
		//}

		if RangeContains(rng, after.Start()) {
			return core.ErrInfo(codes.InvalidArgument,
				"New range overlaps",
				"existing", after,
				"new", rng)
		}
	}

	//if !afterFound {
	//	if beforeFound {
	//		panic("error in badger_map. range after found but not range before")
	//	}
	//} else {
	//	if !beforeFound {
	//		panic("error in badger_map. range before found but not range before")
	//	}
	//
	//	if !RangeEquals(before, after) {
	//		return core.ErrInfo(codes.InvalidArgument,
	//			"Range contains a subset present",
	//			"subset", after,
	//			"new", rng)
	//	}
	//}

	v, err := b.marshal.Marshal(rng)
	if err != nil {
		return status.Convert(err)
	}

	k := rng.Start()
	err = b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(k, v)
	})

	return status.Convert(err)
}

func (b *badgerMap) Remove(r1 Range) (Range, *status.Status) {
	if !ValidRange(r1) {
		return nil, ErrInvalidRange(r1)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	r2, found, stat := b.precedingRange(r1.Start())

	if stat != nil {
		return nil, stat
	}

	if !found {
		return nil, core.ErrInfo(codes.NotFound,
			"Range cannot be deleted. No entries smaller or equal to this",
			"removing", r1)
	}

	if !RangeEquals(r1, r2) {
		return nil, core.ErrInfo(codes.NotFound,
			"Range cannot be deleted because entries dont mismatch",
			"removing", r1,
			"closest", r2)
	}

	err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(r2.Start())
	})
	return r2, status.Convert(err)
}

func NewBadgerRange(name string, marshal RangeEncoder, opt badger.Options) (RangeMap, error) {
	db, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}

	return &badgerMap{
		db:      db,
		name:    name,
		marshal: marshal,
	}, nil
}
