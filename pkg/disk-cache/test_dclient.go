package diskcache

import (
	"context"
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"google.golang.org/grpc"
)

func ExampleUsage() {
	client, err := NewClient(
		context.TODO(),
		"c1",
		"localhost:6001",
		grpc.WithInsecure())
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
