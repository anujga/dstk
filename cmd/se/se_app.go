package main

import (
	"flag"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/sharding_engine/simple"
	"go.uber.org/zap"
)

func main() {
	var port = flag.Int("port", 6001, "grpc port")
	var mode = flag.String("mode", "disk", "sql, disk")
	var connUrl = flag.String("conn", "", "connectionUrl")
	var confPath = flag.String("conf", "./conf", "path of the config folder")
	flag.Parse()

	core.ZapGlobalLevel(zap.InfoLevel)

	var (
		server simple.WorkerAndClient
		err    error
	)

	switch *mode {
	case "disk":
		server, err = simple.UsingLocalFolder(*confPath, true)
	case "sql":
		if len(*connUrl) == 0 {
			panic("connection url missing")
		}
		server, err = simple.UsingSql("postgres", *connUrl)
	}
	if err != nil {
		panic(err)
	}

	f, err := simple.StartServer(*port, server)
	if err != nil {
		panic(err)
	}

	err = f.Wait()
	if err != nil {
		panic(err)
	}

}
