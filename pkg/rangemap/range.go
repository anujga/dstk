package rangemap

import "github.com/anujga/dstk/pkg/ss"

type RangeType byte

const (
	ClosedOpen RangeType = iota
	GreaterThanOrEq
)

type Range struct {
	Start ss.KeyT
	End   ss.KeyT
	Type  RangeType
}

func NewClosedOpenRange(start, end ss.KeyT) *Range {
	return &Range{
		Start: start,
		End:   end,
		Type:  ClosedOpen,
	}
}

func NewGreaterThanRange(start ss.KeyT) *Range {
	return &Range{
		Start: start,
		End:   nil,
		Type:  GreaterThanOrEq,
	}
}
