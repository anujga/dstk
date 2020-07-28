package main

import (
	"flag"
	dc "github.com/anujga/dstk/cmd/disk-cache/core"
	"github.com/anujga/dstk/cmd/disk-cache/gateway"
	"github.com/anujga/dstk/cmd/disk-cache/verify"
	"github.com/anujga/dstk/pkg/core"
	"github.com/posener/complete/v2/compflag"
	"github.com/posener/complete/v2/predict"
	"go.uber.org/zap"
	"gopkg.in/errgo.v2/fmt/errors"
)

func main() {

	var conf = core.MultiStringFlag(
		"conf", "config files. can pass more than 1 --conf c1.yaml --conf c2.yaml")

	var logLevel = zap.LevelFlag(
		"log", zap.InfoLevel, "debug, info, warn, error, dpanic, panic, fatal")

	var mode = compflag.String(
		"mode", "worker", "",
		predict.OptValues("worker", "verify", "gateway"))

	var name = compflag.String("name", "", "eg: dc-1")

	var cleanData = flag.Bool(
		"clean-db", false, "delete existing db")

	compflag.Parse()

	core.ZapGlobalLevel(*logLevel)

	var err error = nil
	c0 := conf.Get()[0]
	switch *mode {
	case "verify":
		err = verify.RunVerifier(c0)
	case "worker":
		var fs []*core.FutureErr

		for _, c := range conf.Get() {
			zap.S().Infow("starting runner", "conf", c)
			f, err := dc.MainRunner(c, *name, *cleanData)
			if err != nil {
				panic(err)
			}
			fs = append(fs, f)
		}

		err = core.Errs(core.WaitMany(fs)...)
	case "gateway":
		c := &gateway.Config{}
		err := core.UnmarshalYaml(c0, c)
		if err != nil {
			break
		}

		f, err := gateway.GatewayMode(c)
		if err != nil {
			break
		}

		err = f.Wait()

	default:
		err = errors.Newf("unknown mode %s", *mode)
	}

	if err != nil {
		panic(err)
	}

}
