package main

import (
	"bytes"
	"context"
	"testing"

	pb "github.com/anujga/dstk/pkg/api/proto"
)

func TestMemStore_Get(t *testing.T) {
	ctx := context.TODO()

	m := NewMemStore(20)

	k := []byte{65}
	v := []byte{66}

	_, err := m.Put(ctx, &pb.DcPutReq{
		Key:        k,
		Value:      v,
		TtlSeconds: 0,
	})

	if err != nil {
		t.Error()
	}

	got, err := m.Get(ctx, &pb.DcGetReq{Key: k})

	if bytes.Compare(v, got.Value) != 0 {
		t.Error()
	}

}
