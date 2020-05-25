package sharder

import (
	"errors"
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"math/rand"
	"sync"
	"time"
)

type ShardStore struct {
	m map[int64]*JobPartitionHolder
}

/*
	For the first partition start = nil
	For the last partition end = nil
*/
type JobPartitionHolder struct {
	id          int64
	t           *rbt.Tree
	index       *TimeIndex
	deleted     *TimeIndex // TODO. How to cleanup deleted ?
	lastPart    *dstk.Partition
	mux         sync.Mutex
	countMux    sync.Mutex
	cyclicCount int16
}

func NewJobPartitionHolder(jobId int64) *JobPartitionHolder {
	jph := JobPartitionHolder{
		id:      jobId,
		t:       rbt.NewWithStringComparator(),
		index:   NewTimeIndex(),
		deleted: NewTimeIndex(),
	}
	return &jph
}

func NewShardStore() *ShardStore {
	ss := ShardStore{m: map[int64]*JobPartitionHolder{}}
	return &ss
}

func (ss *ShardStore) Create(jobId int64, markings [][]byte) error {
	if _, ok := ss.m[jobId]; ok {
		return errors.New("partition for this Job already exist")
	}
	jph := NewJobPartitionHolder(jobId)
	old := []byte(nil)
	for _, cur := range markings {
		partId := jph.generatePartitionId()
		part := dstk.Partition{Id: partId, Start: old, End: cur, Url: jph.getUrl(partId)}
		jph.createPartition(&part)
		old = cur
	}
	partId := jph.generatePartitionId()
	lastPart := dstk.Partition{Id: partId, Start: markings[len(markings)-1], Url: jph.getUrl(partId)}
	jph.createLastPartition(&lastPart)
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

func (ss *ShardStore) Merge(jobId int64, c1 *dstk.Partition, c2 *dstk.Partition) error {
	if jph, ok := ss.m[jobId]; ok {
		return jph.merge(c1, c2)
	}
	return errors.New("invalid job id")
}

func (ss *ShardStore) GetDelta(jobId int64, time int64, activeOnly bool) ([]*dstk.Partition, error) {
	if jph, ok := ss.m[jobId]; ok {
		return jph.getDelta(time, activeOnly), nil
	}
	return nil, errors.New("invalid job id")
}

func (jph *JobPartitionHolder) find(key []byte) *dstk.Partition {
	k := string(key)
	jph.mux.Lock()
	v, found := jph.t.Ceiling(k)
	jph.mux.Unlock()
	if !found {
		return jph.lastPart
	}
	res := v.Value.(*dstk.Partition)
	return res
}

func (jph *JobPartitionHolder) split(marking []byte) error {
	m := string(marking)
	jph.mux.Lock()
	defer jph.mux.Unlock()
	v, found := jph.t.Ceiling(m)
	partition := jph.lastPart
	if found {
		r := v.Value.(*dstk.Partition)
		partition = r
	}
	s := string(partition.GetStart())
	e := string(partition.GetEnd())
	if m == e || m == s {
		return errors.New("invalid marking, partition already exist")
	}
	p1Id := jph.generatePartitionId()
	p2Id := jph.generatePartitionId()
	if s == string(jph.lastPart.GetStart()) {
		p1 := dstk.Partition{Id: p1Id, Start: jph.lastPart.GetStart(), End: marking, Url: jph.getUrl(p1Id)}
		p2 := dstk.Partition{Id: p2Id, Start: marking, Url: jph.getUrl(p2Id)}
		jph.createPartition(&p1)
		jph.replaceLastPartition(&p2)
	} else {
		p1 := dstk.Partition{Id: p1Id, Start: partition.GetStart(), End: marking, Url: jph.getUrl(p1Id)}
		p2 := dstk.Partition{Id: p2Id, Start: marking, End: partition.GetEnd(), Url: jph.getUrl(p2Id)}
		jph.removePartition(partition)
		jph.createPartition(&p1)
		jph.createPartition(&p2)
	}
	return nil
}

func (jph *JobPartitionHolder) merge(c1 *dstk.Partition, c2 *dstk.Partition) error {
	c1s := string(c1.GetStart())
	c1e := string(c1.GetEnd())
	c2s := string(c2.GetStart())
	c2e := string(c2.GetEnd())
	if c1e != c2s {
		return errors.New("non adjacent partitions")
	}
	jph.mux.Lock()
	defer jph.mux.Unlock()
	last := string(jph.lastPart.GetEnd())
	if c2e == last {
		// New partition is the last partition
		p1, f1 := jph.t.Ceiling(c1e)
		if !f1 || string(p1.Value.(*dstk.Partition).GetStart()) != c1s ||
			string(p1.Value.(*dstk.Partition).GetEnd()) != c1e {
			return errors.New("invalid partitions")
		}
		partId := jph.generatePartitionId()
		np := dstk.Partition{
			Id:    partId,
			Start: p1.Value.(*dstk.Partition).GetStart(),
			Url:   jph.getUrl(partId),
		}
		jph.removePartition(p1.Value.(*dstk.Partition))
		jph.replaceLastPartition(&np)
	} else {
		p1, f1 := jph.t.Ceiling(c1e)
		p2, f2 := jph.t.Ceiling(c2e)
		if !f1 || !f2 ||
			string(p1.Value.(*dstk.Partition).GetStart()) != c1s ||
			string(p1.Value.(*dstk.Partition).GetEnd()) != c1e ||
			string(p2.Value.(*dstk.Partition).GetStart()) != c2s ||
			string(p2.Value.(*dstk.Partition).GetEnd()) != c2e {
			return errors.New("invalid partitions")
		}
		partId := jph.generatePartitionId()
		part := dstk.Partition{
			Id:    partId,
			Start: p1.Value.(*dstk.Partition).GetStart(),
			End:   p2.Value.(*dstk.Partition).GetEnd(),
			Url:   jph.getUrl(partId),
		}
		jph.removePartition(p1.Value.(*dstk.Partition))
		jph.removePartition(p2.Value.(*dstk.Partition))
		jph.createPartition(&part)
	}
	return nil
}

func (jph *JobPartitionHolder) getDelta(time int64, activeOnly bool) []*dstk.Partition {
	jph.mux.Lock()
	defer jph.mux.Unlock()
	partitions := make([]*dstk.Partition, 0)
	added := jph.index.ModifiedOnOrAfter(time)
	if added != nil {
		partitions = append(partitions, added...)
	}
	if !activeOnly {
		removed := jph.deleted.ModifiedOnOrAfter(time)
		if removed != nil {
			partitions = append(partitions, removed...)
		}
	}
	return partitions
}

func (jph *JobPartitionHolder) generatePartitionId() int64 {
	return rand.Int63()
}

func (jph *JobPartitionHolder) getUrl(partId int64) string {
	return fmt.Sprintf("jobId: %d--partId: %d--%s", jph.id, partId, "unassigned")
}

func (jph *JobPartitionHolder) createPartition(part *dstk.Partition) {
	part.Active = true
	part.ModifiedOn = jph.timeCounter()
	jph.t.Put(string(part.GetEnd()), part)
	jph.index.Add(part)
}

func (jph *JobPartitionHolder) createLastPartition(part *dstk.Partition) {
	part.Active = true
	part.ModifiedOn = jph.timeCounter()
	jph.lastPart = part
	jph.index.Add(part)
}

func (jph *JobPartitionHolder) removePartition(part *dstk.Partition) {
	jph.t.Remove(string(part.GetEnd()))
	jph.index.Remove(part)

	part.Active = false
	part.ModifiedOn = jph.timeCounter()
	jph.deleted.Add(part)
}

func (jph *JobPartitionHolder) replaceLastPartition(part *dstk.Partition) {
	jph.index.Remove(jph.lastPart)
	jph.lastPart.Active = false
	jph.lastPart.ModifiedOn = jph.timeCounter()
	jph.deleted.Add(jph.lastPart)

	part.Active = true
	part.ModifiedOn = jph.timeCounter()
	jph.lastPart = part
	jph.index.Add(part)
}

func (jph *JobPartitionHolder) timeCounter() int64 {
	jph.countMux.Lock()
	jph.cyclicCount = (jph.cyclicCount + 1) % 1000
	count := time.Now().UnixNano() + int64(jph.cyclicCount)
	jph.countMux.Unlock()
	return count
}
