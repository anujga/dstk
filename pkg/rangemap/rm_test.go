package rangemap

import (
	"github.com/google/btree"
	"testing"
)

func TestRangeMap_Put(t *testing.T) {
	rm := RangeMap{root: btree.New(3)}
	// [abc)defghijklmnopqrstuvwxyz
	err := rm.Put(NewClosedOpenRange([]byte("a"), []byte("c")), "part1")
	if err != nil {
		t.Error(err)
	}
	if got := rm.Get([]byte("b")); got != "part1" {
		t.Errorf("expected %s but got %s", "part1", got)
	}
	// [abc)[cdefg)hijklmnopqrstuvwxyz
	err = rm.Put(NewClosedOpenRange([]byte("c"), []byte("g")), "part2")
	if err != nil {
		t.Error(err)
	}
	if got := rm.Get([]byte("c")); got != "part2" {
		t.Errorf("expected %s but got %s", "part2", got)
	}
	if got := rm.Get([]byte("g")); got != nil {
		t.Errorf("expected nil but got %s", got)
	}
	// [abc)[cdefg)hijklmnopqrstuv[wxyz)
	rng := NewClosedOpenRange([]byte("w"), []byte("z"))
	err = rm.Put(rng, "foobar")
	if err != nil {
		t.Error(err)
	}
	// [abc)[cdefg)hijklmnopqrstuv[wxyz) --NOT OKAY--> [abc)[cdefg)hijkl[mnopqrstuv[wx)yz)
	rng = NewClosedOpenRange([]byte("m"), []byte("x"))
	err = rm.Put(rng, "foobar")
	if err != ErrSuffixOverlaps {
		t.Errorf("Accepted overlapping partitions %v - %v", rng.Start, rng.End)
	}
	// [abc)[cdefg)hijklmnopqrstuv[wxyz) --NOT OKAY--> [abc)[cd[efg)hij)klmnopqrstuv[wxyz)
	rng = NewClosedOpenRange([]byte("e"), []byte("j"))
	err = rm.Put(rng, "foobar")
	if err != ErrPrefixOverlaps {
		t.Errorf("Accepted overlapping partitions %v - %v", rng.Start, rng.End)
	}
	// [abc)[cdefg)hijklmnopqrstuv[wxyz) --NOT OKAY--> [[abc)[cdefg)hijklmnopqrstuv[wxyz))
	rng = NewClosedOpenRange([]byte("a"), []byte("z"))
	err = rm.Put(rng, "foobar")
	if err != ErrPrefixOverlaps {
		t.Errorf("Accepted overlapping partitions %v - %v", rng.Start, rng.End)
	}
	// [abc)[cdefg)hijkl[mnop)qrstuv[wxyz)
	rng = NewClosedOpenRange([]byte("m"), []byte("p"))
	err = rm.Put(rng, "foobar")
	if err != nil {
		t.Error(err)
	}
	// [abc)[cdefg)hijkl[mnop)qrstuv[wxyz) --NOT OKAY--> [abc)[cdefg)hij[kl[mnop)qrs)tuv[wxyz)
	rng = NewClosedOpenRange([]byte("k"), []byte("s"))
	err = rm.Put(rng, "foobar")
	if err != ErrSuffixOverlaps {
		t.Errorf("Accepted overlapping partitions %v - %v", rng.Start, rng.End)
	}
	// partitions now - [abc)[cdefg)hijkl[mnop)qrstuv[wxyz)
	rng = NewGreaterThanRange([]byte("k"))
	err = rm.Put(rng, "foobar")
	if err != ErrSuffixOverlaps {
		t.Errorf("Accepted overlapping partitions %v - %v", rng.Start, rng.End)
	}
	rng = NewGreaterThanRange([]byte("m"))
	err = rm.Put(rng, "foobar")
	if err != ErrPrefixOverlaps {
		t.Errorf("Accepted overlapping partitions %v - %v", rng.Start, rng.End)
	}
	rng = NewGreaterThanRange([]byte("z"))
	err = rm.Put(rng, "zpart")
	if err != nil {
		t.Error(err)
	}
	if got := rm.Get([]byte("zab")); got != "zpart" {
		t.Errorf("expected %s but got %s", "zpart", got)
	}
}
