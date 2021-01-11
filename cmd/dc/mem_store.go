package main

import (
	"context"
	"encoding/base64"
	"sync"

	pb "github.com/anujga/dstk/pkg/api/proto"
)

type Entity struct {
	v      []byte
	expiry int64
}

type memStore struct {
	store map[string]Entity
	mu    sync.Mutex
}

func NewMemStore(defaultSize int32) *memStore {
	return &memStore{store: make(map[string]Entity, defaultSize)}
}

func makeKey(bs []byte) string {
	return base64.StdEncoding.EncodeToString(bs)
}

func (m *memStore) Get(ctx context.Context, req *pb.DcGetReq) (*pb.DcGetRes, error) {
	var v []byte = nil

	k := makeKey(req.Key)

	m.mu.Lock()
	e, found := m.store[k]
	m.mu.Unlock()

	if found {
		v = e.v
	}

	return &pb.DcGetRes{
		Key:   req.Key,
		Value: v,
	}, nil
}

func (m *memStore) Put(ctx context.Context, req *pb.DcPutReq) (*pb.DcRes, error) {
	k := makeKey(req.Key)
	//todo: conflict handling
	m.mu.Lock()

	m.store[k] = Entity{
		v:      req.Value,
		expiry: int64(req.TtlSeconds),
	}
	m.mu.Unlock()

	return &pb.DcRes{}, nil
}

func (m *memStore) Remove(ctx context.Context, req *pb.DcRemoveReq) (*pb.DcRes, error) {
	k := makeKey(req.Key)
	m.mu.Lock()
	delete(m.store, k)
	m.mu.Unlock()
	return &pb.DcRes{}, nil
}
