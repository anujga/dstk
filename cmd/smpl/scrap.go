package main

import (
	"encoding/binary"
	"fmt"
)

func main() {
	x := int64(9223372036854775807)
	x = 102
	b := make([]byte, 64)
	binary.PutVarint(b, x)
	fmt.Println(b)
}
