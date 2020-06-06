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
	"net"
)

type handler func(*Request) (string, error)

type Parts struct {
	Parts []struct {
		Start string
		End   string
	}
}

func addPartitions(partitions *Parts, slog *zap.SugaredLogger, pm *ss.PartitionMgr) error {
	i := 0
	for i, p := range (*partitions).Parts {
		slog.Infow("Adding Partition", "id", i, "end", p)
		pv := dstk.Partition{Id: int64(i), End: []byte(p.End), Start: []byte(p.Start)}
		if err := pm.Add(&pv); err != nil {
			return err
		}
	}
	slog.Infof("partitions count = %d\n", i+1)
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
	parts := new(Parts)
	err = viper.Unmarshal(&parts)
	if err != nil {
		return nil, err
	}
	slog := zap.S()
	slog.Infow("Adding partitions", "keys", parts)
	err = addPartitions(parts, slog, pm)
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
		"conf", "./", "config file")
	flag.Parse()
	viper.AddConfigPath(*conf)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
