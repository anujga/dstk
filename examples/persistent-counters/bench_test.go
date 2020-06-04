//➜ ghz --insecure --proto ./api/protobuf-spec/counter.proto
// --call dstk.CounterRpc.Inc --total=10000 --qps=2000
// -d '{"key":"key1", "value":1}' --connections=50 --concurrency=50
// localhost:9099
package main_test

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"google.golang.org/grpc"
	"testing"
)

func randomKeys(n int) [][]byte {
	rs := make([][]byte, n)
	h := md5.New()
	b := make([]byte, 4)
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint32(b, uint32(i))
		rs[i] = h.Sum(b)
	}
	return rs
}

func BenchmarkFib10(b *testing.B) {
	// run the Fib function b.N times

	ks := randomKeys(100 * 1000)
	N := len(ks)

	for i := 0; i < b.N; i++ {
		b.SetParallelism(6)

		b.RunParallel(func(p *testing.PB) {
			conn, err := grpc.Dial("localhost:9099", grpc.WithInsecure())
			if err != nil {
				b.Fatal(err)
			}
			defer conn.Close()

			c := pb.NewCounterRpcClient(conn)
			for n := 0; n < 1e6; n++ {
				k := ks[n%N]
				c.Inc(context.TODO(), &pb.CounterIncReq{
					Key: string(k), Value: 1,
				})
			}
		})

	}
}
