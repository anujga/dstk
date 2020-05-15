package sharder

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"testing"
)

func TestTimeIndex(t *testing.T) {
	index := NewTimeIndex()
	p0 := dstk.Partition{
		Id:         1,
		Start:      []byte(""),
		End:        []byte("a"),
		Active:     true,
		ModifiedOn: 95,
	}
	p1 := dstk.Partition{
		Id:         1,
		Start:      []byte("a"),
		End:        []byte("h"),
		Active:     true,
		ModifiedOn: 100,
	}
	p2 := dstk.Partition{
		Id:         2,
		Start:      []byte("h"),
		End:        []byte("t"),
		Active:     true,
		ModifiedOn: 100,
	}
	p3 := dstk.Partition{
		Id:         3,
		Start:      []byte("t"),
		End:        []byte("z"),
		Active:     true,
		ModifiedOn: 100,
	}
	index.Add(&p0)
	index.Add(&p1)
	index.Add(&p2)
	index.Add(&p3)
	partitions := index.ModifiedOnOrAfter(90)
	if partitions == nil || len(partitions) != 4 {
		t.Errorf("Partitions count mismatch expected %d, but got %d", 4, len(partitions))
	}
	if !contains(partitions, "", "a", true) || !contains(partitions, "a", "h", true) ||
		!contains(partitions, "h", "t", true) || !contains(partitions, "t", "z", true) {
		t.Error("Expected partitions not found ")
	}
	partitions = index.ModifiedOnOrAfter(100)
	if partitions == nil || len(partitions) != 3 {
		t.Errorf("Partitions count mismatch expected %d, but got %d", 3, len(partitions))
	}
	if !contains(partitions, "h", "t", true) || !contains(partitions, "t", "z", true) ||
		!contains(partitions, "t", "z", true) {
		t.Error("Expected partitions not found ")
	}
	index.Remove(&p2)
	index.Remove(&p2)
	index.Remove(&p2)
	partitions = index.ModifiedOnOrAfter(80)
	if partitions == nil || len(partitions) != 3 {
		t.Errorf("Partitions count mismatch expected %d, but got %d", 3, len(partitions))
	}
	if !contains(partitions, "", "a", true) || !contains(partitions, "a", "h", true) ||
		!contains(partitions, "t", "z", true) {
		t.Error("Expected partitions not found ")
	}
}

//func printPart(partitions []*dstk.Partition) {
//	log.Println("====================")
//	for _, p := range partitions {
//		log.Printf("(%s, %s, %t, %d)\n", p.GetStart(), p.GetEnd(), p.GetActive(), p.GetModifiedOn())
//	}
//}
