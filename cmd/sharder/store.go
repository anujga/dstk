package main

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"go.uber.org/zap"
	"sort"
	"sync"
)

// TODO. Change current naive implementation to an efficient one
type ShardStore struct {
	m map[string]dstk.Partition
	s []string
	mux sync.Mutex
}

func NewShardStore() *ShardStore {
	ss := ShardStore{m: map[string]dstk.Partition{}, s: []string{}}
	return &ss
}

func (ss *ShardStore) Create(p *dstk.Partition) {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	ss.store((*p).GetStart(), *p)
	ss.store((*p).GetEnd(), *p)
}

func (ss *ShardStore) Split(c *dstk.Partition, n1 *dstk.Partition, n2 *dstk.Partition) {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	ss.remove(c.Start)
	ss.store((*n1).GetStart(), *n1)
	ss.store((*n2).GetStart(), *n2)
}

func (ss *ShardStore) Merge(c1 *dstk.Partition, c2 *dstk.Partition, n *dstk.Partition) {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	ss.remove(c1.Start)
	ss.remove(c2.Start)
	ss.store((*n).GetStart(), *n)
}

func (ss *ShardStore) Find(key string) dstk.Partition {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	resKey := ""
	if key == "rccd" {
		zap.L().Info("key = ", zap.String("key", key))
	}
	for i := 0; i < len(ss.s) - 1; i++ {
		start := ss.s[i]
		end := ss.s[i+1]
		if start <= key && key <= end {
			resKey = ss.s[i]
			break
		}
	}
	res :=  dstk.Partition{}
	if resKey != "" {
		res = ss.m[resKey]
	}
	return res
}

func (ss *ShardStore) store(key string, partition dstk.Partition) {
	ss.m[key] = partition
	ss.s = append(ss.s, key)
	sort.Strings(ss.s)
}

func (ss *ShardStore) remove(key string) {
	delete(ss.m, key)
	for i := range ss.s {
		if ss.s[i] == key {
			if i == len(ss.s) - 1 {
				ss.s = ss.s[:i]
			} else {
				ss.s = append(ss.s[:i], ss.s[i+1:]...)
			}
			break
		}
	}
	sort.Strings(ss.s)
}
