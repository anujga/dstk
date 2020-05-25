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

// 2. Define the state for a given partition and implement ss.Consumer
type memCounter struct {
	p    *dstk.Partition
	data map[string]int64
}

func (m *memCounter) Meta() *dstk.Partition {
	return m.p
}

/// this method does not have to be thread safe
func (m *memCounter) Process(msg0 ss.Msg) bool {
	msg := msg0.(*Request)
	v0, found := m.data[msg.K]
	var v1 = msg.V
	if found {
		v1 += v0
	}
	m.data[msg.K] = v1
	return true
}

// 3. implement ss.ConsumerFactory

type memCounterMaker struct {
	maxOutstanding int
}

func (m *memCounterMaker) Make(p *dstk.Partition) (ss.Consumer, int) {
	return &memCounter{
		p:    p,
		data: make(map[string]int64),
	}, m.maxOutstanding
}

// 4. glue it up together
func glue(log *zap.Logger, confPath string) (ss.Router, error) {

	slog := log.Sugar()

	viper.AddConfigPath(confPath)
	viper.ReadInConfig()
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// 4.1 Make the Partition Manager
	factory := &memCounterMaker{viper.GetInt("max-outstanding")}
	pm := ss.NewPartitionMgr(factory, log)

	// 4.2 Register predefined partitions.
	var endParts = 0
	ps := viper.GetStringSlice("parts")
	var i = 0
	for i, p := range ps {
		pv := dstk.Partition{Id: int64(i), End: []byte(p)}
		slog.Info("Adding Part {}. {}", i, pv)
		pm.Add(&pv)
		if len(pv.GetEnd()) == 0 {
			endParts += 1
		}
	}
	slog.Info("partitions count = {}", i+1)

	// 4.3 Ensure presence of end partition
	if endParts != 1 {
		return nil, errors.Newf(
			"exactly 1 end partition required. found: %d", endParts)
	}

	return pm, nil
}

func main() {
	var conf = flag.String(
		"conf", "config.yaml", "config file")
	flag.Parse()

	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	machine, err := glue(log, *conf)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
		if err = machine.OnMsg(&req); err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte("ok"))
	})

	err = http.ListenAndServe(":8080", nil)
	log.Sugar().Fatalw("stopped", "err", err)
}
