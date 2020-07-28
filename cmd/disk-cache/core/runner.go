package dc

import (
	"fmt"
	"github.com/anujga/dstk/cmd/disk-cache/gateway"
	"github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/core/io"
	"github.com/anujga/dstk/pkg/helpers"
	"github.com/anujga/dstk/pkg/sharding_engine"
	"github.com/anujga/dstk/pkg/ss"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/errgo.v2/fmt/errors"
	"os"
)

func MainRunner(conf string, workerName string, cleanDb bool) (*core.FutureErr, error) {
	log := zap.S()

	viper.SetConfigFile(conf)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	chanSize := viper.GetInt64("response_buffer_size")

	log.Infow("parsing workerName",
		"input", workerName)

	wid, err := io.ScaleSetOrdinal(workerName)
	if err != nil {
		return nil, errors.Becausef(nil, err,
			"Invalid workerName. example dstk-2, found %s",
			workerName)
	}

	if wid < 0 {
		return nil, errors.Newf("Bad worker id %s", viper.Get("worker_id"))
	}

	log.Infow("worker id",
		"name", workerName,
		"id", wid)
	workerId := se.WorkerId(wid)
	seUrl := viper.GetString("se_url")
	if cleanDb {
		dbPath := viper.GetString("db_path")
		log.Infow("Cleaning existing db", "path", dbPath)
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

	f := core.RunAsync(func() error {
		metricUrl := viper.GetString("metric_url")
		if metricUrl != "" {
			s := helpers.ExposePrometheus(metricUrl)
			defer core.CloseLogErr(s)
		}

		ws, err := ss.NewWorkerServer(seUrl, workerId, factory)
		if err != nil {
			panic(err)
		}
		dcServer := MakeServer(ws.MsgHandler, chanSize)
		dstk.RegisterDcRpcServer(ws.Server, dcServer)
		err = ws.Start("tcp", url)

		return err

	})

	gw := viper.GetString("GatewayEndpoint")
	if gw != "" {
		clientId := fmt.Sprintf("gw-%d", workerId)
		defer log.Infow("Starting gateway",
			"addr", gw,
			"err", err)

		_, err = gateway.GatewayMode(&gateway.Config{
			Url:      gw,
			SeUrl:    seUrl,
			ClientId: clientId,
		})
		if err != nil {
			return nil, err
		}

	}
	return f, nil
}
