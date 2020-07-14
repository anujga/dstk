package main

import (
	"flag"
	"github.com/anujga/dstk/cmd/disk-cache/verify"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/helpers"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"github.com/anujga/dstk/pkg/ss"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/errgo.v2/fmt/errors"
	"os"
)

func mainRunner(conf string, cleanDb bool) error {
	viper.SetConfigFile(conf)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	chanSize := viper.GetInt64("response_buffer_size")
	wid := viper.GetInt64("worker_id")
	if wid <= 0 {
		return errors.Newf("Bad worker id %s", viper.Get("worker_id"))
	}
	workerId := se.WorkerId(wid)
	targetUrl := viper.GetString("se_url")
	if cleanDb {
		dbPath := viper.GetString("db_path")
		zap.S().Infow("Cleaning existing db", "path", dbPath)
		if err := os.RemoveAll(dbPath); err != nil {
			return err
		}
	}
	factory, err := newConsumerMaker(
		viper.GetString("db_path"),
		viper.GetInt("max_outstanding"))
	if err != nil {
		return err
	}
	f := core.RunAsync(func() error {
		ws, err := ss.NewWorkerServer(targetUrl, workerId, factory, func() interface{} {
			return nil
		})
		if err != nil {
			panic(err)
		}
		dcServer := MakeServer(ws.MsgHandler, chanSize)
		dstk.RegisterDcRpcServer(ws.Server, dcServer)
		return ws.Start("tcp", viper.GetString("url"))
	})
	metrics := helpers.ExposePrometheus(viper.GetString("metric_url"))
	err = f.Wait()
	metrics.Close()
	return err
}

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
		err = mainRunner(*conf, *cleanData)
	}

	if err != nil {
		panic(err)
	}

}
