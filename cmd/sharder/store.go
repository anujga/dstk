package main

import (
	"errors"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"sync"
)

type ShardStore struct {
	m map[int64]*JobPartitionHolder
}

/*
	For the first partition start = nil
	For the last partition end = nil
*/
type JobPartitionHolder struct {
	t        *rbt.Tree
	mux      sync.Mutex
	lastPart *dstk.Partition
}

func NewJobPartitionHolder(last *dstk.Partition) *JobPartitionHolder {
	jph := JobPartitionHolder{t: rbt.NewWithStringComparator(), lastPart: last}
	return &jph
}

func NewShardStore() *ShardStore {
	ss := ShardStore{m: map[int64]*JobPartitionHolder{}}
	return &ss
}

func (ss *ShardStore) Create(jobId int64, partitions []*dstk.Partition, last *dstk.Partition) error {
	if _, ok := ss.m[jobId]; ok {
		return errors.New("partition for this Job already exist")
	}
	jph := NewJobPartitionHolder(last)
	for _, p := range partitions {
		k := string(p.GetEnd())
		jph.t.Put(k, p)
	}
	ss.m[jobId] = jph
	return nil
}

func (ss *ShardStore) Find(jobId int64, key []byte) (*dstk.Partition, error) {
	if jph, ok := ss.m[jobId]; ok {
		return jph.find(key), nil
	}
	return nil, errors.New("invalid job id")
}

func (ss *ShardStore) Split(jobId int64, marking []byte) error {
	if jph, ok := ss.m[jobId]; ok {
		return jph.split(marking)
	}
	return errors.New("invalid job id")
}

func (ss *ShardStore) Merge(c1 *dstk.Partition, c2 *dstk.Partition, n *dstk.Partition) {
	// TODO
}

func (jph *JobPartitionHolder) find(key []byte) *dstk.Partition {
	k := string(key)
	jph.mux.Lock()
	v, found := jph.t.Ceiling(k)
	jph.mux.Unlock()
	if !found {
		return jph.lastPart
	}
	res := v.Value.(dstk.Partition)
	return &res
}

func (jph *JobPartitionHolder) split(marking []byte) error {
	m := string(marking)
	jph.mux.Lock()
	defer jph.mux.Unlock()
	v, found := jph.t.Ceiling(m)
	partition := jph.lastPart
	if found {
		r := v.Value.(dstk.Partition)
		partition = &r
	}
	s := string(partition.GetStart())
	e := string(partition.GetEnd())
	if m == e || m == s {
		return errors.New("invalid marking, partition already exist")
	}
	if s == string(jph.lastPart.GetStart()) {
		part := dstk.Partition{Id: generatePartitionId(), Start: jph.lastPart.Start, End: marking, Url: getUrl()}
		jph.t.Put(marking, &part)
		jph.lastPart.Start = marking
	} else {
		part := dstk.Partition{Id: generatePartitionId(), Start: partition.Start, End: marking, Url: getUrl()}
		jph.t.Put(marking, part)
		partition.Start = marking
	}
	return nil
}
