package main

import (
	"flag"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/errgo.v2/fmt/errors"
)

type handler func(*Request) (string, error)

var log *zap.Logger
var slog *zap.SugaredLogger

func addPartitions(ps []string, slog *zap.SugaredLogger, pm *ss.PartitionMgr) error {
	var endParts = 0
	i := 0
	for i, p := range ps {
		pv := dstk.Partition{Id: int64(i), End: []byte(p)}
		slog.Infow("Adding Partition", "id", i, "end", p)
		if err := pm.Add(&pv); err != nil {
			return err
		}
		if len(pv.GetEnd()) == 0 {
			endParts += 1
		}
	}
	slog.Infof("partitions count = %d\n", i+1)
	// 4.3 Ensure presence of end partition
	if endParts != 1 {
		return errors.Newf(
			"exactly 1 end partition required. found: %d", endParts)
	}
	return nil
}

// 4. glue it up together
func glue() (ss.Router, error) {
	// 4.1 Make the Partition Manager
	factory := &partitionCounterMaker{
		viper.GetString("db_path_prefix"),
		viper.GetInt("max_outstanding"),
	}
	pm := ss.NewPartitionMgr(factory, log)
	// 4.2 Register predefined partitions.
	ps := viper.GetStringSlice("parts")
	err := addPartitions(ps, slog, pm)
	return pm, err
}

// 6. Thick client

func main() {
	router, err := glue()
	if err != nil {
		panic(err)
	}
	rh := ReqHandler{router: router}
	chanSize := viper.GetInt64("response_buffer_size")
	<-server(rh.handle, func() chan interface{} {
		return make(chan interface{}, chanSize)
	})
}

func init() {
	var conf = flag.String(
		"conf", "config.yaml", "config file")
	flag.Parse()
	viper.AddConfigPath(*conf)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	log = zap.L()
	slog = zap.S()
}
