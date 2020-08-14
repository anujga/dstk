package rangemap

import (
	"encoding/binary"
	"encoding/json"
	"github.com/anujga/dstk/pkg/core"
	"github.com/dgraph-io/badger/v2"
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"testing"
)

type TestRange struct {
	KeyStart core.KeyT
	KeyEnd   core.KeyT
	Value    string
}

func trange(start, end, value string) TestRange {
	s := []byte(start)
	if len(s) == 0 {
		s = core.MinKey
	}
	return TestRange{
		KeyStart: s,
		KeyEnd:   []byte(end),
		Value:    value,
	}
}

func (t TestRange) Start() core.KeyT {
	if len(t.KeyStart) == 0 {
		return core.MinKey
	}
	return []byte(t.KeyStart)
}

func (t TestRange) End() core.KeyT {
	return []byte(t.KeyEnd)
}

type TestRangeMarshal struct {
}

func (t *TestRangeMarshal) Marshal(r Range) ([]byte, error) {
	r2 := r.(TestRange)
	return json.Marshal(r2)
}

func (t *TestRangeMarshal) Unmarshal(bytes []byte) (Range, error) {
	a := TestRange{}
	err := json.Unmarshal(bytes, &a)
	return a, err
}

type KeyVal struct {
	key   core.KeyT
	value string
}

type Test struct {
	ranges              []TestRange
	invalidRanges       []TestRange
	keyValues           []KeyVal
	removeValidRanges   []TestRange
	removeInvalidRanges []TestRange
}

func getMaxKey() string {
	i := int64(1024)
	res := make([]byte, 64)
	binary.PutVarint(res, i)
	return string(res)
}

func prepareTests() map[string]Test {
	maxKey := getMaxKey()
	return map[string]Test{
		"simple": {
			ranges: []TestRange{
				trange("a", "o", "H1"),
			},
			invalidRanges: []TestRange{
				trange("d", "k", "H2"),
			},
			keyValues: []KeyVal{
				{[]byte("a"), "H1"},
			},
			removeValidRanges: []TestRange{
				trange("a", "o", "H1"),
			},
			removeInvalidRanges: []TestRange{
				trange("a", "z", "H1"),
			},
		},
		"overlapping": {
			ranges: []TestRange{
				trange("a", "o", "H1"),
				trange("o", "s", "H2"),
				trange("zc", "zz", "H3"),
				trange("", "a", "first"),
				trange("zzz", maxKey, "last"),
			},
			invalidRanges: []TestRange{
				trange("", maxKey, "H1"),
				trange("za", "zze", "H1"),
			},
			keyValues: []KeyVal{
				{[]byte("a"), "H1"},
				{[]byte("o"), "H2"},
				{[]byte("t"), ""},
				{core.MinKey, "first"},
				{[]byte(maxKey), ""},
			},
			removeValidRanges: []TestRange{
				trange("a", "o", "H1"),
				trange("", "a", "first"),
			},
			removeInvalidRanges: []TestRange{
				trange("a", "z", "H1"),
				trange("zzzab", maxKey, "H1"),
			},
		},
	}
}

func TestBtreeRange_Put(t *testing.T) {
	rm := NewBtreeRange(3)
	defer core.CloseLogErr(rm)
	putTests(rm, t)
}
func TestBadgerMap_Put(t *testing.T) {
	dir, err := ioutil.TempDir("", "testBadger")
	if err != nil {
		t.Fatal(err)
	}
	rm, err := NewBadgerRange(
		"TestBadgerMap_Put",
		&TestRangeMarshal{},
		badger.DefaultOptions(dir))

	if err != nil {
		t.Fatal(err)
	}
	defer core.CloseLogErr(rm)
	putTests(rm, t)
}

func putTests(rm RangeMap, t *testing.T) {
	tests := prepareTests()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			for _, rng := range test.ranges {
				if e := rm.Put(rng); e != nil {
					t.Fatalf("Putting range %v failed with error %v", rng, e)
				}
			}
			for _, rng := range test.invalidRanges {
				if e := rm.Put(rng); e == nil {
					t.Fatalf("accepted invalid range %s", rng)
				}
			}
			for _, kv := range test.keyValues {
				var k = []byte(kv.key)
				if len(k) == 0 {
					k = core.MinKey
				}
				rng, found, err := rm.Get(k)
				if err != nil {
					t.Fatal(err)
				}
				if !found {
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
				if _, err := rm.Remove(r); err == nil {
					t.Fatalf("Removed invalid range: %v", r)
				}
			}
		})
	}
}
