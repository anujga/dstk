package main

import (
	"errors"
	"flag"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/sharding_engine/simple"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"math/rand"
)

type Conf struct {
	Port          int
	ConnUrl, Mode string
	Driver        string
	Init          *Bootstrap
}

type Bootstrap struct {
	CleanExisting bool
	NumParts      int
	Seed          int64
	Workers       []*simple.Worker
}

func main() {
	var conf = flag.String("config", "conf.yaml", "conf file")
	core.ZapGlobalLevel(zap.InfoLevel)
	flag.Parse()

	c := &Conf{}
	core.MustUnmarshalYaml(*conf, c)

	f, _, err := run(c)
	if err != nil {
		panic(err)
	}

	err = f.Wait()
	if err != nil {
		panic(err)
	}
	//s.GracefulStop()
}

func run(c *Conf) (*core.FutureErr, *grpc.Server, error) {

	var (
		server simple.WorkerAndClient
		err    error
	)

	switch c.Mode {
	case "sql":
		if len(c.ConnUrl) == 0 {
			return nil, nil, errors.New("connection url missing")
		}
		server, err = simple.UsingSql(c.Driver, c.ConnUrl)
		if err != nil {
			return nil, nil, err
		}

	default:
		return nil, nil, errors.New("sql is the only supported mode")
	}

	init := c.Init
	if init != nil && init.CleanExisting {
		err := bootstrap(c)
		if err != nil {
			return nil, nil, err
		}
	}

	f, s, err := simple.StartServer(c.Port, server)
	if err != nil {
		return nil, nil, err
	}

	return f, s, nil
}

func bootstrap(c *Conf) error {
	init := c.Init

	ps, err := simple.GenerateParts(
		init.NumParts,
		init.Workers,
		rand.NewSource(init.Seed))
	if err != nil {
		return err.Err()
	}

	err2 := simple.InitDb(c.Driver, c.ConnUrl, ps)
	return err2
}
