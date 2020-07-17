package main

import (
	"context"
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	diskcache "github.com/anujga/dstk/pkg/disk-cache"
	"google.golang.org/grpc"
)

func main() {
	client, err := diskcache.NewClient(context.TODO(), "localhost:6001", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	req := &dstk.DcGetReq{Key: []byte("harsha")}
	val, err := client.Get(context.TODO(), req)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(val)
	}
}
