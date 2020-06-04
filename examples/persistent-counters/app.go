package main

import (
	"flag"
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/errgo.v2/fmt/errors"
	"net"
)

type handler func(*Request) (string, error)

func addPartitions(ps []string, slog *zap.SugaredLogger, pm *ss.PartitionMgr) error {
	var endParts = 0
	i := 0
	for i, p := range ps {
		slog.Infow("Adding Partition", "id", i, "end", p)
		pv := dstk.Partition{Id: int64(i), End: []byte(p)}
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
	factory, err := newCounterMaker(
		viper.GetString("db_path"),
		viper.GetInt("max_outstanding"))
	if err != nil {
		return nil, err
	}
	// 4.1 Make the Partition Manager
	pm := ss.NewPartitionMgr(factory, zap.L())
	// 4.2 Register predefined partitions.
	ps := viper.GetStringSlice("parts")
	slog := zap.S()
	slog.Infow("Adding partitions", "keys", ps)
	err = addPartitions(ps, slog, pm)
	return pm, err
}

// 6. Thick client

func startGrpcServer(router ss.Router, log *zap.Logger, resBufSize int64) {
	lis, err := net.Listen("tcp", ":9099")
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}
	s := grpc.NewServer()
	rh := &ss.MsgHandler{Router: router}
	counterServer := MakeServer(rh, log, resBufSize)
	dstk.RegisterCounterRpcServer(s, counterServer)
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		panic(fmt.Sprintf("failed to serve: %v", err))
	}
}

func main() {
	router, err := glue()
	if err != nil {
		panic(err)
	}
	chanSize := viper.GetInt64("response_buffer_size")
	go func() {
		// this is to enable prometheus
		<-server(nil, nil)
	}()
	startGrpcServer(router, zap.L(), chanSize)
}

func init() {
	var conf = flag.String(
		"conf", "config.yaml", "config file")
	flag.Parse()
	viper.AddConfigPath(*conf)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
