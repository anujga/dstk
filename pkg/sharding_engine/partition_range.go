package se

import (
	"fmt"
	"github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/rangemap"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
)

type PartRange struct {
	p *dstk.Partition
}

func (x *PartRange) Start() core.KeyT {
	return x.p.GetStart()
}

func (x *PartRange) End() core.KeyT {
	return x.p.GetEnd()
}

type PartMarshal struct {
}

func (x *PartMarshal) Marshal(r rangemap.Range) ([]byte, error) {
	switch p := r.(type) {
	case *PartRange:
		return proto.Marshal(p.p)
	default:
		return nil, core.ErrInfo(codes.Internal,
			"Bad type given to marshaller",
			"given", fmt.Sprintf("%T", r),
			"allowed", "se.PartRange").Err()
	}
}

func (x *PartMarshal) Unmarshal(data []byte) (rangemap.Range, error) {
	p := &dstk.Partition{}
	err := proto.Unmarshal(data, p)
	if err != nil {
		return nil, err
	}
	return &PartRange{p: p}, nil
}
