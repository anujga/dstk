package verify

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/core/io"
	diskcache "github.com/anujga/dstk/pkg/disk-cache"
	"github.com/anujga/dstk/pkg/helpers"
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

func verifyAll(c *Config) error {
	rpc, err := newClient(c)

	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}

	log := zap.S()

	for i := int64(0); i < c.Count; i++ {
		beg := c.Start + (i * c.Size)
		wg.Add(1)
		go func() {
			defer wg.Done()

			bytes8 := make([]byte, 8)
			for i := int64(0); i < c.Size; i++ {
				uid := beg + i
				binary.LittleEndian.PutUint64(bytes8, uint64(uid))
				uidSer := md5.New().Sum(bytes8)

				res, err := rpc.Get(context.TODO(), &dstk.DcGetReq{Key: uidSer})
				if err != nil {
					log.Errorw("error in get", "err", err)
					//todo: error
				} else {
					document := res.GetDocument()
					etag := document.GetEtag()
					ts := document.GetLastUpdatedEpochSeconds()
					fmt.Printf("Got etag: %s\n, timestamp: %d\n", etag, ts)
					views := binary.LittleEndian.Uint64(document.GetValue())
					expected := uint64(c.Copies) * c.Views
					if views != expected {
						log.Errorw("Mismatch",
							"userId", hex.EncodeToString(uidSer),
							"views", views,
							"expected", expected)
					}
				}
			}
		}()
	}

	wg.Wait()

	return nil
}

func RunVerifier(conf string) error {
	c := &Config{}

	if err := core.UnmarshalYaml(conf, c); err != nil {
		return err
	}

	if c.MetricUrl != "" {
		s := helpers.ExposePrometheus(c.MetricUrl)
		defer core.CloseLogErr(s)
	}

	if err := runMany(c); err != nil {
		return err
	}

	if err := verifyAll(c); err != nil {
		return err
	}

	return nil
}
