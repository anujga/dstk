package rangemap

import (
	"bytes"
	"errors"
	"github.com/anujga/dstk/pkg/core"
	"github.com/google/btree"
)

var ErrInvalidRange = errors.New("invalid range")

type Range interface {
	Start() core.KeyT
	End() core.KeyT
}

type rangeItem struct {
	Range
}

func (r *rangeItem) Less(than btree.Item) bool {
	that := than.(*rangeItem)
	return bytes.Compare(r.Start(), that.Start()) < 0
}

func (r *rangeItem) contains(t core.KeyT) bool {
	return bytes.Compare(r.Start(), t) <= 0 && bytes.Compare(t, r.End()) < 0
}

func (r *rangeItem) precedes(that *rangeItem) bool {
	return bytes.Compare(r.End(), that.Start()) <= 0
}

func NewRange(rng Range) (*rangeItem, error) {
	if rng.End() == nil || bytes.Compare(rng.Start(), rng.End()) < 0 {
		return &rangeItem{rng}, nil
	} else {
		return nil, ErrInvalidRange
	}
}
