package mkv_test

import (
	"context"
	pb "github.com/anujga/dstk/build/gen"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/mkv"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"io/ioutil"
	"net"
	"os"
	"testing"
)

func DataGen1(t *testing.T) string {
	p := pb.MkvPartition{
		Id: 0,
		Entries: []*pb.MkvPartition_Entry{
			{Key: []byte("k1"), Value: []byte("v1")},
			{Key: []byte("k2"), Value: []byte("v2")},
			{Key: []byte("k3"), Value: []byte("v3")},
		},
	}
	bs, err := proto.Marshal(&p)
	f, err := ioutil.TempFile("", "sample_mkv1.pb")
	if err != nil {
		t.Error(err)
	}
	fname := f.Name()
	f.Close()

	if err = ioutil.WriteFile(fname, bs, 0644); err != nil {
		t.Error(err)
	}
	return fname
}

func TestGet(t *testing.T) {
	file := DataGen1(t)
	lis := bufconn.Listen(1024 * 1024)
	s, start := mkv.StartServer(0, lis)
	go start()

	conn, err := grpc.DialContext(
		context.TODO(),
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure())

	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	client := pb.NewMkvClient(conn)

	{
		ex, err := client.AddPart(context.TODO(), &pb.AddParReq{
			Uri: "file://" + file,
		})

		if err != nil {
			t.Error(err)
		}
		if ex.Id != pb.Ex_SUCCESS {
			t.Error(core.WrapEx(ex))
		}

		if err = os.Remove(file); err != nil {
			t.Error(err)
		}
	}

	{
		r, err := client.Get(context.TODO(), &pb.GetReq{
			Key:         []byte("k2"),
			PartitionId: 0,
		})
		if err != nil {
			t.Error(err)
		}
		if r.Ex.Id != pb.Ex_SUCCESS {
			t.Error(core.WrapEx(r.Ex))
		}

		value := string(r.Payload)
		if value != "v2" {
			t.Errorf("Value mismatch. expected %s, recieved %s", "v2", value)
		}
	}
	s.Stop()
}
