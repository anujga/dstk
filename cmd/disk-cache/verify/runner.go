package verify

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	diskcache "github.com/anujga/dstk/pkg/disk-cache"
	"github.com/anujga/dstk/pkg/helpers"
	"github.com/anujga/dstk/pkg/verify"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
	Url, SeUrl string
	MetricUrl  string
	ClientId   string
}

func runMany(c *Config) error {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
	}

	conn, err := grpc.Dial(c.Url, opts...)
	if err != nil {
		return err
	}
	defer core.CloseLogErr(conn)

	//rpc := dstk.NewDcRpcClient(conn)
	rpc, err := diskcache.NewClient(
		context.TODO(),
		c.ClientId,
		c.SeUrl,
		opts...)

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
	conn, err := grpc.Dial(
		c.Url,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
	)
	if err != nil {
		return err
	}
	defer core.CloseLogErr(conn)

	rpc := dstk.NewDcRpcClient(conn)
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
					//todo: error
				} else {
					views := binary.LittleEndian.Uint64(res.GetValue())
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
	s := helpers.ExposePrometheus(c.MetricUrl)

	if err := runMany(c); err != nil {
		return err
	}

	if err := verifyAll(c); err != nil {
		return err
	}

	return s.Close()
}
