package rangemap

import (
	"errors"
	"github.com/anujga/dstk/pkg/core"
	"github.com/google/btree"
)

var (
	ErrRangeOverlaps = errors.New("range overlaps")
	ErrKeyAbsent = errors.New("key absent")
)

type DummyRange struct {
	start core.KeyT
}

func (ki *DummyRange) Start() core.KeyT {
	return ki.start
}

func (ki *DummyRange) End() core.KeyT {
	return nil
}

type RangeMap struct {
	root *btree.BTree
}

func New(degree int) *RangeMap {
	return &RangeMap{root: btree.New(degree)}
}

func (rm *RangeMap) getLessOrEqual(item *rangeItem) *rangeItem {
	var itemInTree *rangeItem
	rm.root.DescendLessOrEqual(item, func(i btree.Item) bool {
		itemInTree = i.(*rangeItem)
		return false
	})
	return itemInTree
}

func (rm *RangeMap) Get(key core.KeyT) (Range, error) {
	item, err := NewRange(&DummyRange{start: key})
	if err != nil {
		return nil, err
	}
	pred := rm.getLessOrEqual(item)
	if pred.contains(key) {
		return pred.Range, nil
	}
	return nil, ErrKeyAbsent
}

func (rm *RangeMap) Put(rng Range) error {
	item, err := NewRange(rng)
	if err != nil {
		return err
	}
	pred := rm.getLessOrEqual(item)
	if pred != nil && !pred.preceeds(item) {
		return ErrRangeOverlaps
	}
	var succ *rangeItem
	rm.root.AscendGreaterOrEqual(item, func(i btree.Item) bool {
		succ = i.(*rangeItem)
		return false
	})
	if succ != nil && !item.preceeds(succ) {
		return ErrRangeOverlaps
	}
	i := rm.root.ReplaceOrInsert(item)
	if i != nil {
		panic("range already exists")
	}
	return nil
}


func (rm *RangeMap) Remove(rng Range) (Range, error) {
	delItem, err := NewRange(rng)
	if err != nil {
		return nil, err
	}
	if item := rm.getLessOrEqual(delItem); item == nil {
		return nil, ErrKeyAbsent
	} else {
		if !item.Less(delItem) && !delItem.Less(item) {
			ri := rm.root.Delete(item)
			return ri.(*rangeItem).Range, nil
		} else {
			return nil, ErrKeyAbsent
		}
	}
}
