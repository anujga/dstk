package core

type KeyT []byte

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
