package partition

import (
	"bytes"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/partition"
	"github.com/google/btree"
	"google.golang.org/grpc/status"
)

type partsItem struct {
	StartBytes []byte
	Parts      map[int64]partition.Actor
}

func (p *partsItem) Less(than btree.Item) bool {
	that := than.(*partsItem)
	return bytes.Compare(p.StartBytes, that.StartBytes) < 0
}

// todo clean up the impl. separate btree logic and business logic
type actorStore struct {
	partRoot     *btree.BTree
	partIdMap    map[int64]partition.Actor
	lastModified int64
}

// Path = control
func (pms *actorStore) add(pa partition.Actor) error {
	var appended bool
	pms.partRoot.DescendLessOrEqual(&partsItem{
		StartBytes: pa.Start(),
	}, func(i btree.Item) bool {
		pi := i.(*partsItem)
		if bytes.Compare(pi.StartBytes, pa.Start()) == 0 {
			appended = true
			pi.Parts[pa.Id()] = pa
		}
		return true
	})
	if !appended {
		pms.partRoot.ReplaceOrInsert(&partsItem{
			StartBytes: pa.Start(),
			Parts:      map[int64]partition.Actor{pa.Id(): pa},
		})
	}
	pms.partIdMap[pa.Id()] = pa
	return nil
}

// Path = control
func (pms *actorStore) remove(pa partition.Actor) (partition.Actor, *status.Status) {
	var delItem *partsItem
	pms.partRoot.DescendLessOrEqual(&partsItem{
		StartBytes: pa.Start(),
	}, func(i btree.Item) bool {
		pi := i.(*partsItem)
		if bytes.Compare(pi.StartBytes, pa.Start()) == 0 {
			delItem = pi
		}
		return true
	})
	if delItem == nil {
		return nil, core.ErrKeyAbsent(pa.Start())
	}
	if len(delItem.Parts) == 1 {
		if delItem.Parts[0] == pa {
			pms.partRoot.Delete(delItem)
		}
	} else {
		delete(delItem.Parts, pa.Id())
	}
	delete(pms.partIdMap, pa.Id())
	return nil, nil
}

// Path = data
func (pms *actorStore) find(key core.KeyT) (partition.Actor, *status.Status) {
	var pa partition.Actor
	pms.partRoot.DescendLessOrEqual(&partsItem{
		StartBytes: key,
	}, func(i btree.Item) bool {
		pi := i.(*partsItem)
		if bytes.Compare(pi.StartBytes, key) <= 0 {
			for _, p := range pi.Parts {
				if bytes.Compare(p.End(), key) > 0 {
					if p.CanServe() {
						pa = p
						return true
					}
				}
			}
		}
		return false
	})
	if pa == nil {
		return nil, core.ErrKeyAbsent(key)
	}
	return pa, nil
}
