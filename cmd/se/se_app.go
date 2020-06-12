package main

import (
	"flag"
	"github.com/anujga/dstk/pkg/sharding_engine/simple"
	"go.uber.org/zap"
)

func main() {
	var port = flag.Int("port", 6001, "grpc port")
	var confPath = flag.String("conf", "./conf", "path of the config folder")
	flag.Parse()

	c := zap.NewDevelopmentConfig()
	c.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	log, err := c.Build()
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(log)

	f, err := simple.StartServer(*port, *confPath)
	if err != nil {
		panic(err)
	}

	err = f.Wait()
	if err != nil {
		panic(err)
	}

}
