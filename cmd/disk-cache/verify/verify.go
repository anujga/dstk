package verify

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/helpers"
	"go.uber.org/zap"
	"sync"
)

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
					views := binary.LittleEndian.Uint64(document.GetValue())
					log.Debugw("Fetched document",
						"Views", views,
						"Etag", document.GetMeta().GetEtag(),
						"Last updated", document.GetMeta().GetLastUpdatedEpochSeconds())

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
		_ = helpers.ExposePrometheus(c.MetricUrl)
	}

	//err := startSplit(c)
	//if err != nil {
	//	return err
	//}

	if err := runMany(c); err != nil {
		return err
	}

	if err := verifyAll(c); err != nil {
		return err
	}

	return nil
}
