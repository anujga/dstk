package main

import (
	"context"
	"fmt"
	diskcache "github.com/anujga/dstk/pkg/disk-cache"
	"google.golang.org/grpc"
)

func main() {
	client, err := diskcache.NewClient(context.TODO(), "localhost:6001", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	val, err := client.Get([]byte("harsha"))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(val)
	}
}
