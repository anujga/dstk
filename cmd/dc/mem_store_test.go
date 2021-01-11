package main

import (
	"bytes"
	"context"
	"os"
	"testing"

	pb "github.com/anujga/dstk/pkg/api/proto"
)

func TestMemStore_Get(t *testing.T) {
	m := NewMemStore(20)
	getput(t, m)
}

func TestBadgerGetPut(t *testing.T) {
	m, err := NewbadgerStore("")
	if err != nil {
		t.Error(err)
	}
	getput(t, m)

	err = os.RemoveAll(m.path)
	if err != nil {
		t.Error(err)
	}
}

func getput(t *testing.T, m pb.DcRpcServer) {
	ctx := context.TODO()
	k := []byte{65}
	v := []byte{66}

	_, err := m.Put(ctx, &pb.DcPutReq{
		Key:        k,
		Value:      v,
		TtlSeconds: 10000,
	})

	if err != nil {
		t.Error(err)
	}

	got, err := m.Get(ctx, &pb.DcGetReq{Key: k})
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(v, got.Value) != 0 {
		t.Error()
	}
}
