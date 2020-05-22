package sharder

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
)

// secondary index exists primarily to quickly answer
// modified on or after query. whenever the there is
// a change in slicer, the clients will get notified
// (or they will pool every few seconds) and they will
// query for all partitions modified since the last
// time they updated
type TimeIndex struct {
	t *rbt.Tree
}

func NewTimeIndex() *TimeIndex {
	return &TimeIndex{
		t: rbt.NewWith(utils.Int64Comparator),
	}
}

func (index *TimeIndex) Add(part *dstk.Partition) {
	m := map[int64]*dstk.Partition(nil)
	res, found := index.t.Get(part.GetModifiedOn())
	if found {
		m = res.(map[int64]*dstk.Partition)
	} else {
		m = map[int64]*dstk.Partition{}
	}
	m[part.GetId()] = part
	index.t.Put(part.GetModifiedOn(), m)
}

func (index *TimeIndex) Remove(p *dstk.Partition) {
	res, found := index.t.Get(p.GetModifiedOn())
	if !found {
		return
	}
	m := res.(map[int64]*dstk.Partition)
	delete(m, p.GetId())
	if len(m) == 0 {
		index.t.Remove(p.GetModifiedOn())
	}
}

func (index *TimeIndex) ModifiedOnOrAfter(time int64) []*dstk.Partition {
	partitions := ([]*dstk.Partition)(nil)
	node, ok := index.t.Ceiling(time)
	if ok {
		partitions = getUpdatedPartitionsFrom(partitions, node)
	}
	return partitions
}

func getUpdatedPartitionsFrom(partitions []*dstk.Partition, node *rbt.Node) []*dstk.Partition {
	partitions = appendPartitions(partitions, node)
	val := partitions[len(partitions)-1].GetModifiedOn()
	next := (*rbt.Node)(nil)
	if node.Right != nil {
		next = node.Right
	} else if node.Parent != nil && node.Parent.Key.(int64) >= val {
		next = node.Parent
	} else {
		return partitions
	}
	return updatePartitions(partitions, node, next, val)
}

func appendPartitions(partitions []*dstk.Partition, node *rbt.Node) []*dstk.Partition {
	m := node.Value.(map[int64]*dstk.Partition)
	for _, v := range m {
		partitions = append(partitions, v)
	}
	return partitions
}

func updatePartitions(partitions []*dstk.Partition, prev *rbt.Node, node *rbt.Node, val int64) []*dstk.Partition {
	if node == nil {
		return partitions
	}
	if node.Left != nil && prev != node.Left && node.Left.Key.(int64) >= val {
		partitions = updatePartitions(partitions, node, node.Left, val)
	}
	partitions = appendPartitions(partitions, node)
	if node.Right != nil && prev != node.Right {
		partitions = updatePartitions(partitions, node, node.Right, val)
	}
	if node.Parent != nil && prev != node.Parent && node.Parent.Key.(int64) >= val {
		partitions = updatePartitions(partitions, node, node.Parent, val)
	}
	return partitions
}
