package dc

import (
	"errors"
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/bdb"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/common"
	"github.com/dgraph-io/badger/v2"
	"go.uber.org/zap"
)

// 2. Define the state for a given partition and implement ss.Consumer
type partitionConsumer struct {
	p      *dstk.Partition
	pc     *bdb.Wrapper
	logger *zap.Logger
	clock  core.DstkClock
}

func (m *partitionConsumer) GetSnapshot() common.AppState {
	return nil
}

func (m *partitionConsumer) ApplySnapshot(as common.AppState) error {
	m.logger.Sugar().Infow("snapshot received", "s", as)
	return nil
}

func (m *partitionConsumer) Meta() *dstk.Partition {
	return m.p
}

// thread safe
func (m *partitionConsumer) get(req *dstk.DcGetReq) (interface{}, error) {
	document, err := m.pc.Get(req.GetKey())
	if err == badger.ErrKeyNotFound {
		return nil, core.ErrKeyAbsent(req.GetKey()).Err()
	}
	return document, err
}

func (m *partitionConsumer) put(req *dstk.DcPutReq) (interface{}, error) {
	meta := &dstk.DcDocumentMeta{
		Etag: req.GetEtag(),
		LastUpdatedEpochSeconds: m.clock.Time(),
	}
	document := &dstk.DcDocument{
		Value: req.GetValue(),
		Meta: meta,
	}
	return nil, m.pc.Put(req.GetKey(), document, req.GetTtlSeconds())
}

func (m *partitionConsumer) remove(req *dstk.DcRemoveReq) (interface{}, error) {
	return nil, m.pc.Remove(req.GetKey())
}

/// this method does not have to be thread safe
func (m *partitionConsumer) Process(msg0 common.Msg) (interface{}, error) {
	msg := msg0.(*DcRequest)
	request := msg.grpcRequest
	switch request.(type) {
	case *dstk.DcGetReq:
		return m.get(request.(*dstk.DcGetReq))
	case *dstk.DcPutReq:
		return m.put(request.(*dstk.DcPutReq))
	case *dstk.DcRemoveReq:
		return m.remove(request.(*dstk.DcRemoveReq))
	default:
		return nil, errors.New(fmt.Sprintf("invalid message %v", msg))
	}
}
