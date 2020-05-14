package main

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"testing"
	"time"
)

func TestShardStore_Create(t *testing.T) {
	store := NewShardStore()
	jobId1 := int64(100)
	marknigs1 := [][]byte{
		[]byte("a"),
		[]byte("j"),
		[]byte("q"),
	}
	jobId2 := int64(101)
	marknigs2 := [][]byte{
		[]byte("c"),
		[]byte("h"),
		[]byte("m"),
		[]byte("t"),
	}
	err := store.Create(jobId1, marknigs1)
	if err != nil {
		t.Error(err)
	}
	err = store.Create(jobId2, marknigs2)
	if err != nil {
		t.Error(err)
	}
	if len(store.m) != 2 {
		t.Errorf("Job count mismatch expected %d, but got %d", 2, len(store.m))
	}
	jph1 := store.m[jobId1]
	jph2 := store.m[jobId2]
	if jph1.t.Size() != 3 {
		t.Errorf("Partitions count mismatch for job id %d, "+
			"expected %d, but got %d", jobId1, 3, jph1.t.Size())
	}
	if jph2.t.Size() != 4 {
		t.Errorf("Partitions count mismatch for job id %d, "+
			"expected %d, but got %d", jobId2, 4, jph2.t.Size())
	}
	if jph1.lastPart == nil {
		t.Errorf("Last partition not set for job id %d, ", jobId1)
	}
	if jph2.lastPart == nil {
		t.Errorf("Last partition not set for job id %d, ", jobId2)
	}
	if string(jph1.lastPart.GetStart()) != string(marknigs1[len(marknigs1)-1]) {
		t.Errorf("Bad start marking for last partition for job id %d, "+
			"expected %s, got %s", jobId1, string(marknigs1[len(marknigs1)-1]),
			string(jph1.lastPart.GetStart()))
	}
	if string(jph2.lastPart.GetStart()) != string(marknigs2[len(marknigs2)-1]) {
		t.Errorf("Bad start marking for last partition for job id %d, "+
			"expected %s, got %s", jobId2, string(marknigs2[len(marknigs2)-1]),
			string(jph2.lastPart.GetStart()))
	}
}

func TestShardStore_GetDelta(t *testing.T) {
	store := NewShardStore()
	jobId := int64(100)
	marknigs := [][]byte{
		[]byte("a"),
		[]byte("h"),
		[]byte("m"),
		[]byte("t"),
	}
	err := store.Create(jobId, marknigs)
	if err != nil {
		t.Error(err)
	}
	curTime := time.Now().UnixNano()
	err = store.Split(jobId, []byte("q"))
	if err != nil {
		t.Error(err)
	}
	partitions, err := store.GetDelta(jobId, curTime, false)
	if err != nil {
		t.Error(err)
	}
	if len(partitions) != 3 {
		t.Errorf("Partitions count mismatch expected %d, "+
			"but got %d", 3, len(partitions))
	}

	if !contains(partitions, "m", "q", true) || !contains(partitions, "q", "t", true) ||
		!contains(partitions, "m", "t", false) {
		t.Error("Expected partitions not found ")
	}
}

func TestShardStore_Split_Part(t *testing.T) {
	store := NewShardStore()
	jobId := int64(100)
	marknigs := [][]byte{
		[]byte("a"),
		[]byte("h"),
		[]byte("m"),
		[]byte("t"),
	}
	err := store.Create(jobId, marknigs)
	if err != nil {
		t.Error(err)
	}
	curTime := time.Now().UnixNano()
	err = store.Split(jobId, []byte("o"))
	if err != nil {
		t.Error(err)
	}
	partitions, err := store.GetDelta(jobId, curTime, false)
	if err != nil {
		t.Error(err)
	}
	if len(partitions) != 3 {
		t.Errorf("Partitions count mismatch expected %d, "+
			"but got %d", 3, len(partitions))
	}

	if !contains(partitions, "m", "o", true) || !contains(partitions, "o", "t", true) ||
		!contains(partitions, "m", "t", false) {
		t.Error("Expected partitions not found ")
	}
}

func TestShardStore_Split_LastPart(t *testing.T) {
	store := NewShardStore()
	jobId := int64(100)
	marknigs := [][]byte{
		[]byte("a"),
		[]byte("h"),
		[]byte("m"),
		[]byte("t"),
	}
	err := store.Create(jobId, marknigs)
	if err != nil {
		t.Error(err)
	}
	curTime := time.Now().UnixNano()
	err = store.Split(jobId, []byte("w"))
	if err != nil {
		t.Error(err)
	}
	partitions, err := store.GetDelta(jobId, curTime, false)
	if err != nil {
		t.Error(err)
	}
	if len(partitions) != 3 {
		t.Errorf("Partitions count mismatch expected %d, "+
			"but got %d", 3, len(partitions))
	}

	if !contains(partitions, "t", "w", true) || !contains(partitions, "w", "", true) ||
		!contains(partitions, "t", "", false) {
		t.Error("Expected partitions not found ")
	}
}

func TestShardStore_Split_FirstPart(t *testing.T) {
	store := NewShardStore()
	jobId := int64(100)
	marknigs := [][]byte{
		[]byte("c"),
		[]byte("h"),
		[]byte("m"),
		[]byte("t"),
	}
	err := store.Create(jobId, marknigs)
	if err != nil {
		t.Error(err)
	}
	curTime := time.Now().UnixNano()
	err = store.Split(jobId, []byte("a"))
	if err != nil {
		t.Error(err)
	}
	partitions, err := store.GetDelta(jobId, curTime, false)
	if err != nil {
		t.Error(err)
	}
	if len(partitions) != 3 {
		t.Errorf("Partitions count mismatch expected %d, "+
			"but got %d", 3, len(partitions))
	}

	if !contains(partitions, "", "a", true) || !contains(partitions, "a", "c", true) ||
		!contains(partitions, "", "c", false) {
		t.Error("Expected partitions not found ")
	}
}

func TestShardStore_Find_Part(t *testing.T) {
	store := NewShardStore()
	jobId := int64(100)
	marknigs := [][]byte{
		[]byte("a"),
		[]byte("h"),
		[]byte("m"),
		[]byte("t"),
	}
	err := store.Create(jobId, marknigs)
	if err != nil {
		t.Error(err)
	}
	err = store.Split(jobId, []byte("o"))
	if err != nil {
		t.Error(err)
	}
	key := "112"
	partition, err := store.Find(jobId, []byte(key))
	if err != nil {
		t.Error(err)
	}

	if !matchPart(partition, "", "a", true) {
		t.Errorf("Find failed for key %s, expected (%s, %s, %t), found (%s, %s, %t)", key,
			"", "a", true,
			partition.GetStart(), partition.GetEnd(), partition.GetActive())
	}

	key = "jack"
	partition, err = store.Find(jobId, []byte(key))
	if err != nil {
		t.Error(err)
	}
	if !matchPart(partition, "h", "m", true) {
		t.Errorf("Find failed for key %s, expected (%s, %s, %t), found (%s, %s, %t)", key,
			"h", "m", true,
			partition.GetStart(), partition.GetEnd(), partition.GetActive())
	}

	key = "omg"
	partition, err = store.Find(jobId, []byte(key))
	if err != nil {
		t.Error(err)
	}
	if !matchPart(partition, "o", "t", true) {
		t.Errorf("Find failed for key %s, expected (%s, %s, %t), found (%s, %s, %t)", key,
			"o", "t", true,
			partition.GetStart(), partition.GetEnd(), partition.GetActive())
	}

	key = "zeebra"
	partition, err = store.Find(jobId, []byte(key))
	if err != nil {
		t.Error(err)
	}
	if !matchPart(partition, "t", "", true) {
		t.Errorf("Find failed for key %s, expected (%s, %s, %t), found (%s, %s, %t)", key,
			"t", "", true,
			partition.GetStart(), partition.GetEnd(), partition.GetActive())
	}
}

func TestShardStore_Merge_Part(t *testing.T) {
	store := NewShardStore()
	jobId := int64(100)
	marknigs := [][]byte{
		[]byte("a"),
		[]byte("h"),
		[]byte("m"),
		[]byte("t"),
	}
	err := store.Create(jobId, marknigs)
	if err != nil {
		t.Error(err)
	}
	partitions, err := store.GetDelta(jobId, 0, false)
	if err != nil {
		t.Error(err)
	}
	if len(partitions) != 5 {
		t.Errorf("Partitions count mismatch expected %d, "+
			"but got %d", 5, len(partitions))
	}
	curTime := time.Now().UnixNano()
	c1 := dstk.Partition{
		Start: []byte("h"),
		End:   []byte("m"),
	}
	c2 := dstk.Partition{
		Start: []byte("m"),
		End:   []byte("t"),
	}
	err = store.Merge(jobId, &c1, &c2)
	if err != nil {
		t.Error(err)
	}
	partitions, err = store.GetDelta(jobId, curTime, false)
	if err != nil {
		t.Error(err)
	}
	if len(partitions) != 3 {
		t.Errorf("Partitions count mismatch expected %d, "+
			"but got %d", 3, len(partitions))
	}

	if !contains(partitions, "h", "m", false) || !contains(partitions, "m", "t", false) ||
		!contains(partitions, "h", "t", true) {
		t.Error("Expected partitions not found ")
	}
	partitions, err = store.GetDelta(jobId, 0, true)
	if err != nil {
		t.Error(err)
	}
	if len(partitions) != 4 {
		t.Errorf("Partitions count mismatch expected %d, "+
			"but got %d", 4, len(partitions))
	}

	if !contains(partitions, "", "a", true) || !contains(partitions, "a", "h", true) ||
		!contains(partitions, "h", "t", true) || !contains(partitions, "t", "", true) {
		t.Error("Expected partitions not found ")
	}
}

func contains(partitions []*dstk.Partition, s string, e string, active bool) bool {
	for _, p := range partitions {
		if matchPart(p, s, e, active) {
			return true
		}
	}
	return false
}

func matchPart(p *dstk.Partition, s string, e string, active bool) bool {
	return string(p.GetStart()) == s && string(p.GetEnd()) == e && p.GetActive() == active
}
