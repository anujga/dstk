package main

import (
	"context"
	"flag"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"github.com/anujga/dstk/pkg/ss"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

//type Parts struct {
//	Parts []struct {
//		Start string
//		End   string
//	}
//}

//func addPartitions(partitions *Parts, slog *zap.SugaredLogger, pm *ss.PartitionMgr) error {
//	i := 0
//	for i, p := range (*partitions).Parts {
//		slog.Infow("Adding Partition", "id", i, "end", p)
//		pv := dstk.Partition{Id: int64(i), End: []byte(p.End), Start: []byte(p.Start)}
//		if err := pm.Add(&pv); err != nil {
//			return err
//		}
//	}
//	slog.Infof("partitions count = %d\n", i+1)
//	return nil
//}

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
	// 4.2 Register predefined partitions.
	//parts := new(Parts)
	//err = viper.Unmarshal(&parts)
	//if err != nil {
	//	return nil, err
	//}
	//slog := zap.S()
	//slog.Infow("Adding partitions", "keys", parts)
	//err = addPartitions(parts, slog, pm)
	return pm, err
}

// 6. Thick client

func startGrpcServer(router ss.Router, log *zap.Logger, resBufSize int64, rh *ss.MsgHandler) error {
	lis, err := net.Listen("tcp", ":9099")
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	cacheServer := MakeServer(rh, log, resBufSize)
	dstk.RegisterDcRpcServer(s, cacheServer)
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		return err
	}
	return nil
}

func main() {
	chanSize := viper.GetInt64("response_buffer_size")
	workerId := se.WorkerId(viper.GetInt64("worker_id"))
	targetUrl := viper.GetString("seServer")
	rpc, err := se.NewSeWorker(context.TODO(), targetUrl)
	if err != nil {
		panic(err)
	}

	router, err := glue(workerId, rpc)
	if err != nil {
		panic(err)
	}

	f := core.RunAsync(func() error {
		msgHandler := &ss.MsgHandler{Router: router}
		return startGrpcServer(router, zap.L(), chanSize, msgHandler)
	})
	err = f.Wait()
	if err != nil {
		panic(err)
	}
}

func init() {
	var conf = flag.String(
		"conf", "./cmd/disk-cache", "config file")
	flag.Parse()
	viper.AddConfigPath(*conf)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
