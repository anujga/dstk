package core

import "time"

type KeyT []byte
type IdGenerator func() int64

var (
	MaxKey KeyT
	MinKey KeyT
)

const maxKeyLen = 1024

func init() {
	MinKey = []byte{0}
	MaxKey = make([]byte, maxKeyLen+1)
	for i := 0; i < maxKeyLen; i++ {
		MaxKey[i] = 0xff
	}
	MaxKey[maxKeyLen] = 1
}

func ValidKey(k KeyT) bool {
	n := len(k)
	//todo: maxKeyLen check is weak. we are allowing keys of size larger than MaxKey
	return n > 0 && n < maxKeyLen+1
}

type DstkClock interface {
	Time() int64
}

type RealClock struct {
}

func (r *RealClock) Time() int64 {
	return time.Now().Unix()
}

type EtagGenerator interface {
	Next(curr int64) int64
	Initial() int64
}

type SequentialEtagGenerator struct {
}

func (s *SequentialEtagGenerator) Next(curr int64) int64 {
	return curr + 1
}

func (s *SequentialEtagGenerator) Initial() int64 {
	return 1
}
