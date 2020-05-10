package slicer

import (
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"sync"
)

//todo: implement using immutable tree and get rid of mutex
type rbtHolder struct {
	t        *rbt.Tree
	mu       sync.Mutex
	lastPart Partition
}

func StaticHolder(parts []Partition, lastPart Partition) SliceRdr {
	r := &rbtHolder{
		t:        rbt.NewWithStringComparator(),
		lastPart: lastPart,
	}

	for _, p := range parts {
		k := string(p.End())
		r.t.Put(k, p)
	}

	return r
}

func ReadPartsFrom(filename string) []Partition {
	return nil
}

func (r *rbtHolder) Get(key []byte) (Partition, error) {
	k := string(key)
	r.mu.Lock()
	v, found := r.t.Floor(k)
	r.mu.Unlock()
	if !found {
		return r.lastPart, nil
	}

	return v.Value.(Partition), nil
}
