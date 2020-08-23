package rangemap

import (
	"fmt"
	"github.com/anujga/dstk/pkg/core"
	"github.com/google/btree"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BtreeRange struct {
	root *btree.BTree
}

func (rm *BtreeRange) Close() error {
	rm.root.Clear(false)
	return nil
}

func NewBtreeRange(degree int) RangeMap {
	return &BtreeRange{root: btree.New(degree)}
}

func (rm *BtreeRange) getLessOrEqual(item *rangeItem) *rangeItem {
	var itemInTree *rangeItem
	rm.root.DescendLessOrEqual(item, func(i btree.Item) bool {
		itemInTree = i.(*rangeItem)
		return false
	})
	return itemInTree
}

func (rm *BtreeRange) Iter(start core.KeyT) <-chan Range {
	item := NewKeyRange(start)
	ch := make(chan Range)
	go func() {
		rm.root.AscendGreaterOrEqual(item, func(i btree.Item) bool {
			i0 := i.(*rangeItem)
			ch <- i0.Range
			return true
		})
		close(ch)
	}()

	return ch
}

func (rm *BtreeRange) Get(key core.KeyT) (Range, bool, *status.Status) {
	item, err := NewRange(&dummyRange{start: key})
	if err != nil {
		return nil, false, status.Convert(err)
	}
	pred := rm.getLessOrEqual(item)
	if pred == nil {
		return nil, false, nil
	}
	if RangeContains(pred, key) {
		return pred.Range, true, nil
	}
	return nil, false, nil
}

func (rm *BtreeRange) Put(rng Range) *status.Status {
	item, err := NewRange(rng)
	if err != nil {
		return status.Convert(err)
	}
	pred := rm.getLessOrEqual(item)
	if pred != nil && !pred.precedes(item) {
		return status.Convert(
			fmt.Errorf("%v overlaps with %v", item, pred))
	}
	var succ *rangeItem
	rm.root.AscendGreaterOrEqual(item, func(i btree.Item) bool {
		succ = i.(*rangeItem)
		return false
	})
	if succ != nil && !item.precedes(succ) {
		return status.Convert(
			fmt.Errorf("%v overlaps with %v", item, succ))
	}
	i := rm.root.ReplaceOrInsert(item)
	if i != nil {
		return status.Convert(fmt.Errorf("range %v already exists", rng))
	}
	return nil
}

func (rm *BtreeRange) Remove(rng Range) (Range, *status.Status) {
	delItem, err := NewRange(rng)
	if err != nil {
		return nil, status.Convert(err)
	}
	if item := rm.getLessOrEqual(delItem); item == nil {
		return nil, ErrInvalidRange(rng)
	} else {
		if !item.Less(delItem) && !delItem.Less(item) {
			ri := rm.root.Delete(item)
			return ri.(*rangeItem).Range, nil
		} else {
			return nil, core.ErrInfo(
				codes.NotFound,
				"potential match a is not b",
				"potential", item,
				"requested", delItem)
		}
	}
}
