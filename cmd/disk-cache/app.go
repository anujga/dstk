package main

import (
	"flag"
	dc "github.com/anujga/dstk/cmd/disk-cache/core"
	"github.com/anujga/dstk/cmd/disk-cache/verify"
	"github.com/anujga/dstk/pkg/core"
	"go.uber.org/zap"
)

func main() {
	var conf = flag.String(
		"conf", "config.yaml", "config file")

	var logLevel = zap.LevelFlag(
		"log", zap.InfoLevel, "debug, info, warn, error, dpanic, panic, fatal")

	var verifyFlag = flag.Bool(
		"verify", false, "run program in verification mode")

	var cleanData = flag.Bool(
		"clean-db", false, "delete existing db")

	flag.Parse()

	core.ZapGlobalLevel(*logLevel)

	var err error = nil
	if *verifyFlag {
		err = verify.RunVerifier(*conf)
	} else {
		err = dc.MainRunner(*conf, *cleanData)
	}

	if err != nil {
		panic(err)
	}

}
