package main

import (
	"flag"
	dc "github.com/anujga/dstk/cmd/disk-cache/core"
	"github.com/anujga/dstk/cmd/disk-cache/verify"
	"github.com/anujga/dstk/pkg/core"
	"go.uber.org/zap"
)

func main() {

	var conf = core.MultiStringFlag(
		"conf", "config files. can pass more than 1 --conf c1.yaml --conf c2.yaml")

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
		err = verify.RunVerifier(conf.Get()[0])
	} else {
		var fs []*core.FutureErr

		for _, c := range conf.Get() {
			zap.S().Infow("starting runner", "conf", c)
			f, err := dc.MainRunner(c, *cleanData)
			if err != nil {
				panic(err)
			}
			fs = append(fs, f)

		}

		err = core.Errs(core.WaitMany(fs)...)
	}

	if err != nil {
		panic(err)
	}

}
