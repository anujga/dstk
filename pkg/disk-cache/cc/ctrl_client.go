package main

import (
	"context"
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, _ := grpc.DialContext(context.TODO(), "localhost:6011", grpc.WithInsecure())
	client := dstk.NewWorkerCtrlClient(conn)
	res, err := client.SplitPartition(context.TODO(), &dstk.SplitPartReq{
		SourcePartition:  &dstk.Partition{
			Id:         0,
			Start:      []byte("a"),
			End:        []byte("o"),
		},
		TargetPartitions: &dstk.Partitions{Parts: []*dstk.Partition{
			{
				Id:         101,
				Start:      []byte("a"),
				End:        []byte("e"),
			},
			{
				Id:         102,
				Start:      []byte("e"),
				End:        []byte("o"),
			},
		}},
	})
	fmt.Println(res)
	fmt.Println(err)
}
