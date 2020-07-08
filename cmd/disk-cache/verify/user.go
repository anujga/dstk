package verify

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"errors"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/verify"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"math/rand"
	"sync"
	"time"
)

type user struct {
	id    int64
	idSer []byte

	totalViews         uint64
	maxViewsPerSession int64
	ttlSeconds         float32

	rnd    rand.Source
	rpc    dstk.DcRpcClient
	bytes8 []byte
}

type ProcessFactory = func(id int64) *user

func NewUserFactory(totalViews uint64, rpc dstk.DcRpcClient) ProcessFactory {
	bytes8 := make([]byte, 8)
	fn := func(id int64) *user {
		binary.LittleEndian.PutUint64(bytes8, uint64(id))

		return &user{
			id:                 id,
			idSer:              md5.New().Sum(bytes8),
			totalViews:         totalViews,
			maxViewsPerSession: 5,
			ttlSeconds:         float32(1 * time.Minute),
			rnd:                rand.NewSource(id),
			rpc:                rpc,
			bytes8:             bytes8,
		}
	}

	return fn
}

func (u *user) Invoke(ctx context.Context) error {
	if u.Done(ctx) {
		return errors.New("Called after done")
	}
	var v1 = 1 + uint64(u.rnd.Int63()%u.maxViewsPerSession)
	if v1 > u.totalViews {
		v1 = u.totalViews
	}

	res, err := u.rpc.Get(ctx, &dstk.DcGetReq{Key: u.idSer})
	if err != nil {
		return err
	}

	v0 := binary.LittleEndian.Uint64(res.Value)
	v2 := v0 + v1
	binary.LittleEndian.PutUint64(u.bytes8, v2)

	_, err = u.rpc.Put(ctx, &dstk.DcPutReq{
		Key:        u.idSer,
		Value:      u.bytes8,
		TtlSeconds: u.ttlSeconds,
	})
	if err != nil {
		return err
	}

	u.totalViews -= v1
	return nil
}

func (u *user) Done(context.Context) bool {
	return u.totalViews > 0
}

func CreateUsers(beg int64, n int64, fn ProcessFactory) ([]verify.Process, error) {
	if n < 1 {
		return nil, errors.New("Invalid Arg, beg - end > 0")
	}

	ps := make([]verify.Process, n)
	for i := int64(0); i < n; i++ {
		ps[i] = fn(i + beg)
	}

	return ps, nil

}

type Config struct {
	Start  int64
	Size   int64
	Count  int64
	Seed   int64
	Views  uint64
	Copies int64
	Url    string
}

func RunMany(c *Config) error {
	conn, err := grpc.Dial(c.Url, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer core.CloseLogErr(conn)

	rpc := dstk.NewDcRpcClient(conn)
	fn := NewUserFactory(c.Views, rpc)

	wg := &sync.WaitGroup{}

	for i := int64(0); i < c.Count; i++ {
		beg := c.Start + (i * c.Size)
		rnd := rand.NewSource(c.Seed)
		wg.Add(1)
		go func() {
			defer wg.Done()
			ps, err := CreateUsers(beg, c.Size, fn)
			if err != nil {
				zap.S().Error("Error creating users", "err", err)
				return
			}

			verify.RunProcess(&verify.SampledProcess{
				Ps:  ps,
				Rnd: rnd,
			})

		}()
	}

	wg.Wait()

	return nil
}

func VerifyAll(c *Config) error {
	conn, err := grpc.Dial(c.Url, grpc.WithInsecure())
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
						log.Error("Mismatch",
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
