package sharder

import (
	"errors"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"go.uber.org/zap"
	"sync"
)

type ClientShardStore struct {
	m            map[int64]*ClientJobPartitionHolder
	ci           *client
	stop         chan bool
	jobs         []int64
	lastModified int64
}

type ClientJobPartitionHolder struct {
	id       int64
	t        *rbt.Tree
	lastPart *dstk.Partition
	mux      sync.Mutex
}

func NewClientJobPartitionHolder(jobId int64) *ClientJobPartitionHolder {
	cjph := ClientJobPartitionHolder{
		id: jobId,
		t:  rbt.NewWithStringComparator(),
	}
	return &cjph
}

func NewClientShardStore(jobs []int64) *ClientShardStore {
	css := ClientShardStore{
		m:    map[int64]*ClientJobPartitionHolder{},
		ci:   getClientInfo(),
		stop: make(chan bool),
		jobs: jobs,
	}
	return &css
}

func (store *ClientShardStore) TrackJob(jobId int64) {
	if store.jobs == nil {
		store.jobs = make([]int64, 0)
	}
	if _, ok := store.m[jobId]; ok {
		return
	}
	cjph := NewClientJobPartitionHolder(jobId)
	store.m[jobId] = cjph
	go store.initStore(jobId)
}

func (store *ClientShardStore) Update(jobId int64, added []*dstk.Partition, removed []*dstk.Partition) {
	logger := zap.L()
	cjph := store.m[jobId]
	if cjph == nil {
		logger.Error("Invalid job id: ", zap.Any("id", jobId))
		return
	}
	cjph.mux.Lock()
	for _, part := range removed {
		cjph.t.Remove(part.GetEnd())
	}
	lastModified := store.lastModified
	for _, part := range added {
		if part.GetEnd() == nil {
			cjph.lastPart = part
		} else {
			cjph.t.Put(part.GetEnd(), part.GetEnd())
		}
		cjph.t.Put(part.GetEnd(), part)
		if part.GetModifiedOn() > lastModified {
			lastModified = part.GetModifiedOn()
		}
	}
	store.lastModified = lastModified
	cjph.mux.Unlock()
}

func (store *ClientShardStore) Find(jobId int64, key []byte) (*dstk.Partition, error) {
	logger := zap.L()
	cjph := store.m[jobId]
	if cjph == nil {
		logger.Error("Invalid job id: ", zap.Any("id", jobId))
		return nil, errors.New("invalid job id")
	}
	res, found := cjph.t.Ceiling(key)
	if found {
		return res.Value.(*dstk.Partition), nil
	}
	return cjph.lastPart, nil
}
