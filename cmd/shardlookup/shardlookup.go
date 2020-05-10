package main
//
//import (
//	"context"
//	"go.uber.org/zap"
//	"log"
//	"time"
//
//	dstk "github.com/anujga/dstk/pkg/api/proto"
//	"google.golang.org/grpc"
//)
//
//const (
//	address = "127.0.0.1:50000"
//)
//
//func initLogger() {
//	logger, _ := zap.NewProduction()
//	zap.ReplaceGlobals(logger)
//}
//
//type client struct {
//	conn   *grpc.ClientConn
//	c      *dstk.ShardStorageClient
//	ctx    *context.Context
//	cancel context.CancelFunc
//}
//
//func getClientInfo() *client {
//	logger := zap.L()
//	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
//	if err != nil {
//		logger.Error("Connect failed: ", zap.Error(err))
//	}
//
//	c := dstk.NewShardStorageClient(conn)
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//	return &client{conn: conn, c: &c, ctx: &ctx, cancel: cancel}
//}
//
//func main() {
//	initLogger()
//	logger := zap.L()
//	defer syncLogs(logger)
//
//	ci := getClientInfo()
//	defer checkClose(ci.conn)
//	defer ci.cancel()
//
//	createPartition(ci, "aaaa", "zzzz")
//	findPartition(ci, "aaaa")
//	findPartition(ci, "kdbd")
//	findPartition(ci, "zzzz")
//	splitPartition(ci, "aaaa", "mmmm", "zzzz")
//	findPartition(ci, "aaaa")
//	findPartition(ci, "cccc")
//	findPartition(ci, "jzzz")
//	findPartition(ci, "rccd")
//	splitPartition(ci, "aaaa", "gggg", "mmmm")
//	findPartition(ci, "jzzz")
//	findPartition(ci, "rccd")
//	mergePartition(ci, "gggg", "mmmm", "mmmm", "zzzz")
//	findPartition(ci, "jzzz")
//}
//
//func createPartition(ci *client, start string, end string) {
//	logger := zap.L()
//	client := *ci.c
//	p1 := dstk.Partition{Id: 1, Start: start, End: end}
//	req := dstk.CreateReq{C: &p1}
//	res, err := client.CreatePartition(*ci.ctx, &req)
//	if err != nil {
//		logger.Error("CreatePartition Failed: ", zap.Error(err))
//	} else {
//		logger.Info("CreatePartition Success: ", zap.Any("Response", res))
//	}
//}
//
//func splitPartition(ci *client, start string, marker string, end string) {
//	logger := zap.L()
//	client := *ci.c
//	c := dstk.Partition{Id: 1, Start: start, End: end}
//	n1 := dstk.Partition{Id: 1, Start: start, End: marker}
//	n2 := dstk.Partition{Id: 1, Start: marker, End: end}
//	req := dstk.SplitReq{C: &c, N1: &n1, N2: &n2}
//	res, err := client.SplitPartition(*ci.ctx, &req)
//	if err != nil {
//		logger.Error("SplitPartition Failed: ", zap.Error(err))
//	} else {
//		logger.Info("SplitPartition Success: ", zap.Any("Response", res))
//	}
//}
//
//func mergePartition(ci *client, s1 string, e1 string, s2 string, e2 string) {
//	logger := zap.L()
//	client := *ci.c
//	c1 := dstk.Partition{Id: 1, Start: s1, End: e1}
//	c2 := dstk.Partition{Id: 1, Start: s2, End: e2}
//	n := dstk.Partition{Id: 1, Start: s1, End: e2}
//	req := dstk.MergeReq{C1: &c1, C2: &c2, N: &n}
//	res, err := client.MergePartition(*ci.ctx, &req)
//	if err != nil {
//		logger.Error("MergePartition Failed: ", zap.Error(err))
//	} else {
//		logger.Info("MergePartition Success: ", zap.Any("Response", res))
//	}
//}
//
//func findPartition(ci *client, key string) {
//	logger := zap.L()
//	client := *ci.c
//	findRes, err := client.FindPartition(*ci.ctx, &dstk.Find_Req{Key: key})
//	if err != nil {
//		logger.Error("FindPartition Failed: ", zap.Error(err))
//	} else {
//		logger.Info("FindPartition Success: ",
//			zap.String("key", key),
//			zap.String("start", findRes.Par.GetStart()),
//			zap.String("end", findRes.Par.GetEnd()))
//	}
//}
//
//func syncLogs(logger *zap.Logger) {
//	err := logger.Sync()
//	if err != nil {
//		log.Fatal("Error syncing log!", err)
//	}
//}
//
//func checkClose(conn *grpc.ClientConn) {
//	logger := zap.L()
//	err := conn.Close()
//	if err != nil {
//		logger.Error("Error closing connection!", zap.Error(err))
//	}
//}
