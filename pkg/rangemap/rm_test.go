package rangemap

import (
	"encoding/binary"
	"github.com/google/go-cmp/cmp"
	"testing"
)

type TestRange struct {
	KeyStart []byte
	KeyEnd   []byte
	Value    string
}

func (t TestRange) Start() []byte {
	return t.KeyStart
}

func (t TestRange) End() []byte {
	return t.KeyEnd
}

type KeyVal struct {
	key   []byte
	value string
}

type Test struct {
	ranges              []TestRange
	invalidRanges       []TestRange
	keyValues           []KeyVal
	removeValidRanges   []TestRange
	removeInvalidRanges []TestRange
}

func getMaxKey() []byte {
	i := int64(1024)
	res := make([]byte, 64)
	binary.PutVarint(res, i)
	return res
}

func prepareTests() map[string]Test {
	maxKey := getMaxKey()
	return map[string]Test{
		"simple": {
			ranges: []TestRange{
				{[]byte("a"), []byte("o"), "H1"},
			},
			invalidRanges: []TestRange{
				{[]byte("d"), []byte("k"), "H2"},
			},
			keyValues: []KeyVal{
				{key: []byte("a"), value: "H1"},
			},
			removeValidRanges: []TestRange{
				{[]byte("a"), []byte("o"), "H1"},
			},
			removeInvalidRanges: []TestRange{
				{[]byte("a"), []byte("z"), "H1"},
			},
		},
		"overlapping": {
			ranges: []TestRange{
				{[]byte("a"), []byte("o"), "H1"},
				{[]byte("o"), []byte("s"), "H2"},
				{[]byte("zc"), []byte("zz"), "H3"},
				{[]byte(""), []byte("a"), "first"},
				{[]byte("zzz"), maxKey, "last"},
			},
			invalidRanges: []TestRange{
				{[]byte(""), maxKey, "H1"},
				{[]byte("za"), []byte("zze"), "H1"},
			},
			keyValues: []KeyVal{
				{key: []byte("a"), value: "H1"},
				{key: []byte("o"), value: "H2"},
				{key: []byte("t"), value: ""},
				{key: []byte(""), value: "first"},
				{key: maxKey, value: ""},
			},
			removeValidRanges: []TestRange{
				{[]byte("a"), []byte("o"), "H1"},
				{[]byte(""), []byte("a"), "first"},
			},
			removeInvalidRanges: []TestRange{
				{[]byte("a"), []byte("z"), "H1"},
				{[]byte("zzzab"), maxKey, "H1"},
			},
		},
	}
}

func TestRangeMap_Put(t *testing.T) {
	tests := prepareTests()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			rm := New(3)
			for _, rng := range test.ranges {
				if e := rm.Put(rng); e != nil {
					t.Fatalf("Putting range %v failed with error %v", rng, e)
				}
			}
			for _, rng := range test.invalidRanges {
				if e := rm.Put(rng); e != ErrRangeOverlaps {
					t.Fatalf("accepted invalid range %s", rng)
				}
			}
			for _, kv := range test.keyValues {
				rng, err := rm.Get(kv.key)
				if err == ErrKeyAbsent {
					if kv.value != "" {
						t.Fatalf("failed to get value for %v", rng)
					}
				} else {
					testRange := rng.(TestRange)
					if diff := cmp.Diff(kv.value, testRange.Value); diff != "" {
						t.Fatalf(diff)
					}
				}
			}
			for _, r := range test.removeValidRanges {
				if removed, err := rm.Remove(r); err != nil {
					t.Fatal(err)
				} else {
					if diff := cmp.Diff(r, removed); diff != "" {
						t.Fatalf(diff)
					}
				}
			}
			for _, r := range test.removeInvalidRanges {
				if _, err := rm.Remove(r); err != ErrKeyAbsent {
					t.Fatalf("Removed invalid range: %v", r)
				}
			}
		})
	}
}
