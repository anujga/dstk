package main

import (
	"context"
	"flag"
	"github.com/anujga/dstk/cmd/disk-cache/verify"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"github.com/anujga/dstk/pkg/ss"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
<<<<<<< HEAD
	"google.golang.org/grpc/reflection"
	"gopkg.in/errgo.v2/fmt/errors"
	"net"
	"os"
=======
>>>>>>> Support partition split
)

// 4. glue it up together
func glue(workerId se.WorkerId, rpc dstk.SeWorkerApiClient) (ss.Router, error) {
	factory, err := newConsumerMaker(
		viper.GetString("db_path"),
		viper.GetInt("max_outstanding"))
	if err != nil {
		return nil, err
	}
	// 4.1 Make the Partition Manager
	pm := ss.NewPartitionMgr2(workerId, factory, rpc)

	return pm, err
}

<<<<<<< HEAD
// 6. Thick client

func startGrpcServer(url string, resBufSize int64, rh *ss.MsgHandler) error {
	lis, err := net.Listen("tcp", url)
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	cacheServer := MakeServer(rh, resBufSize)
	dstk.RegisterDcRpcServer(s, cacheServer)
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		return err
	}
	return nil
}

func mainRunner(conf string, cleanDb bool) error {
	viper.SetConfigFile(conf)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

=======
func main() {
	core.ZapGlobalLevel(zap.InfoLevel)
>>>>>>> Support partition split
	chanSize := viper.GetInt64("response_buffer_size")
	wid := viper.GetInt64("worker_id")
	if wid <= 0 {
		return errors.Newf("Bad worker id %s", viper.Get("worker_id"))
	}
	workerId := se.WorkerId(wid)
	seUrl := viper.GetString("se_url")
	factory, err := newConsumerMaker(
		viper.GetString("db_path"),
		viper.GetInt("max_outstanding"))

	rpc, err := se.NewSeWorker(context.TODO(), targetUrl, grpc.WithInsecure())
	if err != nil {
		return err
	}

	if cleanDb {
		dbPath := viper.GetString("db_path")
		zap.S().Infow("Cleaning existing db", "path", dbPath)
		if err := os.RemoveAll(dbPath); err != nil {
			return err
		}
	}
	router, err := glue(workerId, rpc)
	if err != nil {
		return err
	}

	f := core.RunAsync(func() error {
<<<<<<< HEAD
		msgHandler := &ss.MsgHandler{Router: router}
		return startGrpcServer(viper.GetString("url"), chanSize, msgHandler)
=======
		ws, err := ss.NewWorkerServer(seUrl, workerId, factory, zap.L())
		if err != nil {
			panic(err)
		}
		return ws.Start()
		return ss.StartServer(func(server *grpc.Server, msh *ss.MsgHandler) {
			cacheServer := MakeServer(msh, zap.L(), chanSize)
			dstk.RegisterDcRpcServer(server, cacheServer)
		})
>>>>>>> Support partition split
	})
	err = f.Wait()
	if err != nil {
		return err
	}
	return nil
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
