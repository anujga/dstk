package rangemap

import (
	"bytes"
	"github.com/anujga/dstk/pkg/core"
	"github.com/google/btree"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrInvalidRange(r Range) *status.Status {
	return core.ErrInfo(
		codes.InvalidArgument,
		"invalid range",
		"start", r.Start(),
		"end", r.End())
}

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
		return nil, ErrInvalidRange(rng).Err()
	}
}

func NewKeyRange(k core.KeyT) *rangeItem {
	return &rangeItem{&dummyRange{start: k}}
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
