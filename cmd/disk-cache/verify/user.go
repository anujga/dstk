package verify

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/verify"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"time"
)

type user struct {
	id    	int64
	idSer 	[]byte
	copyId 	int

	totalViews         uint64
	maxViewsPerSession int64
	ttlSeconds         float32

	rnd    rand.Source
	rpc    dstk.DcRpcClient
	bytes8 []byte
}

type ProcessFactory = func(userId int64, copyId int) *user

func NewUserFactory(totalViews uint64, rpc dstk.DcRpcClient) ProcessFactory {
	fn := func(userId int64, copyId int) *user {
		bytes8 := make([]byte, 8)
		binary.LittleEndian.PutUint64(bytes8, uint64(userId))
		idSer := md5.New().Sum(bytes8)
		zap.S().Debugw("Creating user", "id", userId, "idSer", idSer)
		return &user{
			id:                 userId,
			idSer:              idSer,
			totalViews:         totalViews,
			maxViewsPerSession: 5,
			ttlSeconds:         float32(5 * time.Minute),
			rnd:                rand.NewSource(userId),
			rpc:                rpc,
			bytes8:             bytes8,
			copyId:		    copyId,
		}
	}
	return fn
}

func (u *user) Invoke(ctx context.Context) error {
	log := zap.S()
	if u.Done(ctx) {
		return errors.New("Called after done")
	}
	var delta = 1 + uint64(u.rnd.Int63()%u.maxViewsPerSession)
	if delta > u.totalViews {
		delta = u.totalViews
	}

	res, err := u.rpc.Get(ctx, &dstk.DcGetReq{Key: u.idSer})

	var v0 = uint64(0)
	if err != nil {
		e0 := status.Convert(err)
		if e0.Code() == codes.NotFound {
			log.Debugw("Adding new key",
				"uid", hex.EncodeToString(u.idSer),
				"copy", u.copyId)
		} else {
			return err
		}
	} else {
		v0 = binary.LittleEndian.Uint64(res.GetDocument().GetValue())
	}
	v2 := v0 + delta

	log.Debugw("Doing put request",
		"uid", hex.EncodeToString(u.idSer),
		"copy", u.copyId,
		"existing", v0,
		"new", v2,
		"views remaining", u.totalViews - delta)

	binary.LittleEndian.PutUint64(u.bytes8, v2)
	_, err = u.rpc.Put(ctx, &dstk.DcPutReq{
		Key:        u.idSer,
		Value:      u.bytes8,
		TtlSeconds: u.ttlSeconds,
		Etag:       res.GetDocument().GetEtag(),
	})
	if err != nil {
		return err
	}
	u.totalViews -= delta
	return nil
}

func (u *user) Done(context.Context) bool {
	return u.totalViews == 0
}

func (u *user) Init(ctx context.Context) error {
	log := zap.S()
	if u.Done(ctx) {
		return errors.New("Called after done")
	}
	// We will only run init for the first copy of each user.
	if u.copyId != 0 {
		log.Debugw("Not resetting values for",
			"uid", hex.EncodeToString(u.idSer),
			"copy", u.copyId)
		return nil
	}
	resetValue := uint64(0)
	log.Debugw("Resetting",
		"uid", hex.EncodeToString(u.idSer),
		"copy", u.copyId,
		"reset value", resetValue)
	resetBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(resetBytes, resetValue)
	_, err := u.rpc.Put(ctx, &dstk.DcPutReq{
		Key:        u.idSer,
		Value:      resetBytes,
		TtlSeconds: u.ttlSeconds,
	})
	return err
}

func CreateUsers(beg int64, n int64, copyId int, fn ProcessFactory) ([]verify.Process, error) {
	if n < 1 {
		return nil, errors.New("Invalid Arg, beg - end > 0")
	}

	ps := make([]verify.Process, n)
	for i := int64(0); i < n; i++ {
		ps[i] = fn(i + beg, copyId)
	}
	return ps, nil
}
