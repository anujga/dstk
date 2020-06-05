package rangemap

import (
	"bytes"
	"fmt"
	"github.com/anujga/dstk/pkg/ss"
	"github.com/google/btree"
)

type RangeItemType byte

const (
	Start RangeItemType = iota
	End
	Point
)

type RangeItem struct {
	key       ss.KeyT
	itemType  RangeItemType
	rangeType RangeType
	value     interface{}
}

func (r *RangeItem) Less(than btree.Item) bool {
	that := than.(*RangeItem)
	compVal := bytes.Compare(r.key, that.key)
	if compVal == 0 {
		return isThisLess(r, that)
	}
	return compVal < 0
}

func isThisLess(this *RangeItem, that *RangeItem) bool {
	if this.rangeType == that.rangeType {
		return compareSameRangeType(this, that)
	} else if this.itemType == Point || that.itemType == Point {
		return comparePoint(this, that)
	} else if this.rangeType == GreaterThanOrEq || that.rangeType == GreaterThanOrEq {
		return compareGreaterThan(this, that)
	} else {
		panic(fmt.Sprintf("unknown range types in %v or %v", this, that))
	}
}

func compareSameRangeType(this, that *RangeItem) bool {
	if this.itemType == that.itemType {
		return false
	} else {
		switch this.rangeType {
		case ClosedOpen:
			// open end is lesser than closed end
			return this.itemType == End
		case GreaterThanOrEq:
			panic("greater-than can not have more than one item type")
		default:
			panic("unknown range type")
		}
	}
}

func comparePoint(this, that *RangeItem) bool {
	// assumes one and only one of them is a point range
	nonPt := that
	if that.itemType == Point {
		nonPt = this
	}
	switch nonPt.rangeType {
	case ClosedOpen:
		if nonPt.itemType == Start {
			// both are equal
			return false
		} else {
			// if non point node corresponds to end of closed open, then non point node is lesser
			return nonPt == this
		}
	case GreaterThanOrEq:
		// point and greater-than range that have same key are equal
		return false
	default:
		panic("unknown range type")
	}
}

func compareGreaterThan(this, that *RangeItem) bool {
	// assumes one and only one of them is a greater-than range
	nonGte := that
	if that.rangeType == GreaterThanOrEq {
		nonGte = this
	}
	switch nonGte.rangeType {
	case ClosedOpen:
		if nonGte.itemType == End {
			return nonGte == this
		}
		return false
	default:
		panic("invalid range type combinations")
	}
}
