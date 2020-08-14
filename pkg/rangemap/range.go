package rangemap

import (
	"bytes"
	"fmt"
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

type RangeEncoder interface {
	Marshal(r Range) ([]byte, error)
	Unmarshal([]byte) (Range, error)
}

func RangeContains(r Range, t core.KeyT) bool {
	return bytes.Compare(r.Start(), t) <= 0 && bytes.Compare(t, r.End()) < 0
}

func RangeEquals(r1 Range, r2 Range) bool {
	startEq := bytes.Compare(r1.Start(), r2.Start()) == 0
	if !startEq {
		return false
	}
	endEq := bytes.Compare(r1.End(), r2.End()) == 0
	return endEq
}

//todo: deprecate the rangeItem wrapper
type rangeItem struct {
	Range
}

func (r *rangeItem) Less(than btree.Item) bool {
	that := than.(*rangeItem)
	return bytes.Compare(r.Start(), that.Start()) < 0
}

func (r *rangeItem) precedes(that *rangeItem) bool {
	return bytes.Compare(r.End(), that.Start()) <= 0
}
func ValidRange(r Range) bool {
	return r.End() == nil || bytes.Compare(r.Start(), r.End()) < 0
}

func NewRange(rng Range) (*rangeItem, error) {
	if ValidRange(rng) {
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

func (ki *dummyRange) String() string {
	return fmt.Sprintf("type=dummy, start=%v", ki.start)
}

func (ki *dummyRange) Start() core.KeyT {
	return ki.start
}

func (ki *dummyRange) End() core.KeyT {
	return nil
}
