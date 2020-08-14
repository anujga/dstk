package verify

import (
	"context"
	"github.com/anujga/dstk/pkg/actions/split"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core/io"
	diskcache "github.com/anujga/dstk/pkg/disk-cache"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"github.com/anujga/dstk/pkg/verify"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gopkg.in/errgo.v2/fmt/errors"
	"math/rand"
	"sync"
)

type Config struct {
	Start      int64
	Size       int64
	Count      int64
	Seed       int64
	Views      uint64
	Copies     int64
	SeUrl      string
	GatewayUrl string
	Mode       string
	MetricUrl  string
	ClientId   string
}

func newClient(c *Config) (dstk.DcRpcClient, error) {
	zap.S().Infow("New client",
		"mode", c.Mode)
	switch c.Mode {
	default:
		return nil, errors.Newf("Bad mode %s", c.Mode)
	case "gateway":
		opts := io.DefaultClientOpts()
		conn, err := grpc.DialContext(
			context.TODO(), c.GatewayUrl, opts...)
		if err != nil {
			return nil, err
		}

		return dstk.NewDcRpcClient(conn), nil
	case "client":
		return diskcache.NewClient(
			context.TODO(),
			c.ClientId,
			c.SeUrl,
			io.DefaultClientOpts()...)
	}
}

func runMany(c *Config) error {

	rpc, err := newClient(c)

	if err != nil {
		return err
	}

	//time.Sleep(5 * time.Second)
	fn := NewUserFactory(c.Views, rpc)

	wg := &sync.WaitGroup{}

	for i := int64(0); i < c.Count; i++ {
		beg := c.Start + (i * c.Size)
		rnd := rand.NewSource(c.Seed)
		wg.Add(1)
		go func() {
			defer wg.Done()
			ps, err := CreateUsers(beg, c.Size, fn)
			log := zap.S()
			if err != nil {
				log.Error("Error creating users", "err", err)
				return
			}

			stats := verify.RunProcess(&verify.SampledProcess{
				Ps:  ps,
				Rnd: rnd,
			})

			log.Infow("stats",
				"beg", beg,
				"stats", stats)

		}()
	}

	wg.Wait()

	return nil
}

func startSplit(c *Config) error {
	seRpc, err := se.NewSeClient(context.TODO(), c.SeUrl)
	if err != nil {
		return err
	}
	partsRes, err := seRpc.GetPartitions(context.TODO(), &dstk.PartitionGetRequest{FetchAll: true})
	if err != nil {
		return err
	}
	ps := partsRes.GetPartitions().GetParts()
	rnd := rand.NewSource(c.Seed)
	i := int64(len(ps))
	splitDag := split.Dag{
		Part: ps[rnd.Int63()%int64(len(ps))],
		IdGenerator: func() int64 {
			i = i + 1
			return i
		},
		SeRpc: seRpc,
	}
	return splitDag.Start(context.TODO(), nil)

}
