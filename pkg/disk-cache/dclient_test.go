package diskcache

import (
	"context"
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"google.golang.org/grpc"
	"testing"
)

var k = []byte{0xa}

func put(client dstk.DcRpcClient) {
	res, err := client.Put(context.TODO(), &dstk.DcPutReq{
		Key:        k,
		Value:      []byte("harsha"),
		TtlSeconds: 1000,
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res)
	}
}

func getClient() dstk.DcRpcClient {
	client, err := NewClient(
		context.TODO(),
		"c1",
		"localhost:6001",
		grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return client
}

func get(client dstk.DcRpcClient) {
	req := &dstk.DcGetReq{Key: k}
	val, err := client.Get(context.TODO(), req)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(val)
	}
}

func TestExampleUsage(t *testing.T) {
	c := getClient()
	put(c)
	get(c)
	//fmt.Println(hex.DecodeString("aa"))
}
