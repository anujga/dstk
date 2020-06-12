package core

type KeyT []byte

var MinKey, MaxKey KeyT

const MaxKeyLen = 1024

func init() {
	MinKey = []byte("")
	MaxKey = make([]byte, MaxKeyLen+1)
	MaxKey[MaxKeyLen] = 1
}
