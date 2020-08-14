package core

import "time"

type KeyT []byte
type IdGenerator func()int64

var (
	MaxKey KeyT
	MinKey KeyT
)

const MaxKeyLen = 1024

func init() {
	MinKey = []byte("")
	MaxKey = make([]byte, MaxKeyLen+1)
	for i := 0; i < MaxKeyLen; i++ {
		MaxKey[i] = 0xff
	}
	MaxKey[MaxKeyLen] = 1
}


type DstkClock interface {
	Time() int64
}

type RealClock struct {
}

func (r *RealClock) Time() int64 {
	return time.Now().Unix()
}
