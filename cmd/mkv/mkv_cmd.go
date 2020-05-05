package main

import (
	"flag"
	"github.com/anujga/dstk/pkg/mkv"
)

func main() {
	var port = flag.Int("port", 6001, "grpc port")
	flag.Parse()
	_, blockingStart := mkv.StartServer(int32(*port), nil)
	blockingStart()
}
