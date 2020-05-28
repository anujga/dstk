package main

import (
	"encoding/json"
	"flag"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/errgo.v2/fmt/errors"
	"net/http"
	"time"
)

// 1. Define the Request payload and implement the ss.Msg interface
type Request struct {
	K string
	V int64
}

func (i *Request) ReadOnly() bool {
	return false
}

func (i *Request) Key() ss.KeyT {
	return []byte(i.K)
}

// 4. glue it up together
func glue(confPath string) (ss.Router, error) {

	log := zap.L()
	slog := zap.S()

	viper.AddConfigPath(confPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// 4.1 Make the Partition Manager
	factory := &partitionCounterMaker{viper.GetString("db_path_prefix"), viper.GetInt("max_outstanding")}
	pm := ss.NewPartitionMgr(factory, log)

	// 4.2 Register predefined partitions.
	var endParts = 0
	ps := viper.GetStringSlice("parts")
	var i = 0
	for i, p := range ps {
		pv := dstk.Partition{Id: int64(i), End: []byte(p)}
		slog.Infow("Adding Partition", "id", i, "end", p)
		if err := pm.Add(&pv); err != nil {
			return nil, err
		}
		if len(pv.GetEnd()) == 0 {
			endParts += 1
		}
	}
	slog.Infof("partitions count = %d\n", i+1)

	// 4.3 Ensure presence of end partition
	if endParts != 1 {
		return nil, errors.Newf(
			"exactly 1 end partition required. found: %d", endParts)
	}

	return pm, nil
}

type handler func(*Request) (string, error)

// 5. server for partitions
func server(callback handler) chan error {
	var err error

	http.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		var req Request
		var err error

		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
		payload, err := callback(&req)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte(payload))
	})

	ch := make(chan error, 1)
	go func() {
		err = http.ListenAndServe(":8080", nil)
		zap.S().Infow("stopped", "err", err)
		ch <- err
	}()

	return ch
}

// 6. Thick client

//
func main() {
	var conf = flag.String(
		"conf", "config.yaml", "config file")
	flag.Parse()
	router, err := glue(*conf)
	if err != nil {
		panic(err)
	}
	servingFuture := server(func(msg *Request) (string, error) {
		// TODO: make sync/async handling configurable
		if channel, err := router.OnMsg(msg); err != nil {
			return "", err
		} else {
			select {
			case e := <-channel:
				if e == nil {
					return "ok", nil
				} else {
					return "internal error", e
				}
			case _ = <-time.After(time.Second * 5):
				return "internal error", errors.New("timedout")
			}
		}
	})
	<-servingFuture
}