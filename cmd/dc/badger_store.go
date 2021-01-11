package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"go.uber.org/zap"

	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/bdb"
)

//type Entity struct {
//	v      []byte
//	expiry int64
//}

type badgerStore struct {
	pc   *bdb.Wrapper
	path string
}

func NewbadgerStore(path string) (*badgerStore, error) {
	var err error

	if path == "" {
		path, err = ioutil.TempDir("", "testBadger")
		if err != nil {
			return nil, err
		}
	}

	zap.S().Infow("Badger disk location", zap.String("path", path))
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}

	opt := badger.DefaultOptions(path).
		WithTableLoadingMode(options.LoadToRAM).
		WithValueLogLoadingMode(options.MemoryMap).
		WithSyncWrites(false)

	db, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}

	return &badgerStore{
		pc:   &bdb.Wrapper{DB: db},
		path: path,
	}, nil
}

func (m *badgerStore) Get(ctx context.Context, req *pb.DcGetReq) (*pb.DcGetRes, error) {
	v, err := m.pc.Get(req.Key)
	if err != nil {
		return nil, err
	}

	return &pb.DcGetRes{
		Key:   req.Key,
		Value: v,
	}, nil
}

func (m *badgerStore) Put(ctx context.Context, req *pb.DcPutReq) (*pb.DcRes, error) {

	err := m.pc.Put(req.Key, req.Value, req.TtlSeconds)
	if err != nil {
		return nil, err
	}
	return &pb.DcRes{}, nil
}

func (m *badgerStore) Remove(ctx context.Context, req *pb.DcRemoveReq) (*pb.DcRes, error) {
	err := m.pc.Remove(req.Key)
	if err != nil {
		return nil, err
	}
	return &pb.DcRes{}, nil
}
