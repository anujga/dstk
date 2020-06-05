package rangemap

import (
	"errors"
	"fmt"
	"github.com/anujga/dstk/pkg/ss"
	"github.com/google/btree"
)


var (
	ErrPrefixOverlaps   = errors.New("range prefix overlaps")
	ErrSuffixOverlaps   = errors.New("range suffix overlaps")
)

type RangeMap struct {
	root *btree.BTree
}

func (rm *RangeMap) getLessOrEqual(item *RangeItem) *RangeItem {
	var itemInTree *RangeItem
	rm.root.DescendLessOrEqual(item, func(i btree.Item) bool {
		itemInTree = i.(*RangeItem)
		return false
	})
	return itemInTree
}

func (rm *RangeMap) Get(key ss.KeyT) interface{} {
	item := &RangeItem{key: key, itemType: Point, rangeType: ClosedOpen, value: nil}
	treeItem := rm.getLessOrEqual(item)
	if treeItem.itemType == Start {
		return treeItem.value
	} else {
		return nil
	}
}

func getItemsForRange(rng *Range, value interface{}) (*RangeItem, *RangeItem) {
	startItem := &RangeItem{key: rng.Start, value: value, itemType: Start, rangeType: rng.Type}
	var endItem *RangeItem
	switch rng.Type {
	case ClosedOpen:
		endItem = &RangeItem{key: rng.End, itemType: End, rangeType: rng.Type}
	case GreaterThanOrEq:
		endItem = nil
	default:
		panic(fmt.Sprintf("unknown range type %v", rng.Type))
	}
	return startItem, endItem
}

func (rm *RangeMap) checkValidStart(startItem *RangeItem) error {
	pred := rm.getLessOrEqual(startItem)
	if pred != nil {
		switch pred.itemType {
		case Start:
			// not valid for range types we support
			return ErrPrefixOverlaps
		case End:
			// we support closed-open ranges next to each other
			if !pred.Less(startItem) {
				return ErrPrefixOverlaps
			}
		default:
			panic("invalid node found in tree")
		}
	}
	return nil
}

func (rm *RangeMap) checkValidSuccessor(startItem, endItem *RangeItem, rngType RangeType) error {
	var successor *RangeItem
	rm.root.AscendGreaterOrEqual(startItem, func(i btree.Item) bool {
		successor = i.(*RangeItem)
		return false
	})
	if successor != nil {
		switch rngType {
		case GreaterThanOrEq:
			return ErrSuffixOverlaps
		case ClosedOpen:
			if successor.itemType == End {
				// This should never happen
				return ErrSuffixOverlaps
			} else if !endItem.Less(successor) {
				return ErrSuffixOverlaps
			}
		}
	}
	return nil
}

func (rm *RangeMap) Put(rng *Range, value interface{}) error {
	startItem, endItem := getItemsForRange(rng, value)
	if err := rm.checkValidStart(startItem); err != nil {
		return err
	}
	if err := rm.checkValidSuccessor(startItem, endItem, rng.Type); err != nil {
		return err
	}
	i := rm.root.ReplaceOrInsert(startItem)
	if i != nil {
		panic("incorrect validation")
	}
	if endItem != nil {
		i = rm.root.ReplaceOrInsert(endItem)
		if i != nil {
			panic("incorrect impossible")
		}
	}
	return nil
}

//func (rm *RangeMap) Remove(rng Range) {
//	rm.root.Delete(rm.Get(rng.Start))
//}
