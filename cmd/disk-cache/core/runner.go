package dc

import (
	"github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/helpers"
	"github.com/anujga/dstk/pkg/sharding_engine"
	"github.com/anujga/dstk/pkg/ss"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/errgo.v2/fmt/errors"
	"os"
)

func MainRunner(conf string, cleanDb bool) (*core.FutureErr, error) {
	viper.SetConfigFile(conf)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	chanSize := viper.GetInt64("response_buffer_size")
	wid := viper.GetInt64("worker_id")
	if wid <= 0 {
		return nil, errors.Newf("Bad worker id %s", viper.Get("worker_id"))
	}
	workerId := se.WorkerId(wid)
	targetUrl := viper.GetString("se_url")
	if cleanDb {
		dbPath := viper.GetString("db_path")
		zap.S().Infow("Cleaning existing db", "path", dbPath)
		if err := os.RemoveAll(dbPath); err != nil {
			return nil, err
		}
	}
	factory, err := newConsumerMaker(
		viper.GetString("db_path"),
		viper.GetInt("max_outstanding"))
	if err != nil {
		return nil, err
	}

	url := viper.GetString("url")

	metricUrl := viper.GetString("metric_url")
	if metricUrl != "" {
		s := helpers.ExposePrometheus(metricUrl)
		defer core.CloseLogErr(s)
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
		err = ws.Start("tcp", url)
		var err2 error = nil
		//err2 := metrics.Close()
		return core.Errs(err, err2)

	})
	return f, nil
}
