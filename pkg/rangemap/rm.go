package rangemap

import (
	"fmt"
	"github.com/anujga/dstk/pkg/core"
	"github.com/google/btree"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrKeyAbsent(k core.KeyT) *status.Status {
	return core.ErrInfo(
		codes.InvalidArgument,
		"key absent",
		"key", k)
}

type dummyRange struct {
	start core.KeyT
}

func (ki *dummyRange) Start() core.KeyT {
	return ki.start
}

func (ki *dummyRange) End() core.KeyT {
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

func (rm *RangeMap) Iter(start core.KeyT) chan interface{} {
	item := NewKeyRange(start)
	ch := make(chan interface{})
	go func() {
		rm.root.AscendGreaterOrEqual(item, func(i btree.Item) bool {
			ch <- i
			return true
		})
		close(ch)
	}()

	return ch
}

func (rm *RangeMap) Get(key core.KeyT) (Range, error) {
	item, err := NewRange(&dummyRange{start: key})
	if err != nil {
		return nil, err
	}
	pred := rm.getLessOrEqual(item)
	if pred.contains(key) {
		return pred.Range, nil
	}
	return nil, ErrKeyAbsent(key).Err()
}

func (rm *RangeMap) Put(rng Range) error {
	item, err := NewRange(rng)
	if err != nil {
		return err
	}
	pred := rm.getLessOrEqual(item)
	if pred != nil && !pred.precedes(item) {
		return fmt.Errorf("%v overlaps with %v", item, pred)
	}
	var succ *rangeItem
	rm.root.AscendGreaterOrEqual(item, func(i btree.Item) bool {
		succ = i.(*rangeItem)
		return false
	})
	if succ != nil && !item.precedes(succ) {
		return fmt.Errorf("%v overlaps with %v", item, succ)
	}
	i := rm.root.ReplaceOrInsert(item)
	if i != nil {
		return fmt.Errorf("range %v already exists", rng)
	}
	return nil
}

func (rm *RangeMap) Remove(rng Range) (Range, error) {
	delItem, err := NewRange(rng)
	if err != nil {
		return nil, err
	}
	if item := rm.getLessOrEqual(delItem); item == nil {
		return nil, ErrInvalidRange(rng).Err()
	} else {
		if !item.Less(delItem) && !delItem.Less(item) {
			ri := rm.root.Delete(item)
			return ri.(*rangeItem).Range, nil
		} else {
			return nil, fmt.Errorf("potential match %v is not %v", item, delItem)
		}
	}
}
