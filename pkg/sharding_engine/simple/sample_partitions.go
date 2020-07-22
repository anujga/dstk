package simple

import (
	"encoding/binary"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"time"
)

type Worker struct {
	Id  int64
	Url string
}

func GenerateParts(m int, ws []*Worker, src rand.Source) ([]*pb.Partition, *status.Status) {
	maxParts := 1 << 15
	if m > maxParts {
		return nil, core.ErrInfo(
			codes.InvalidArgument,
			"GenerateParts cannot create so many parts",
			"given", m,
			"maxAllowed", maxParts)
	}

	max16 := 1 << 16
	n := max16 / m
	ps := make([]*pb.Partition, 0, m)
	rnd := rand.New(src)
	for i := 0; i < m; i++ {
		j := rnd.Intn(len(ws))
		w := ws[j]
		beg := i * n
		p := newPart(int16(beg), int16(beg+n), w)
		ps = append(ps, p)
	}

	ps[m-1].End = core.MaxKey
	return ps, nil
}

func newPart(beg, end int16, w *Worker) *pb.Partition {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(beg))
	e := make([]byte, 2)
	binary.BigEndian.PutUint16(e, uint16(end))

	p := &pb.Partition{
		Id:         int64(beg),
		ModifiedOn: time.Now().UnixNano(),
		Active:     true,
		Start:      b,
		End:        e,
		Url:        w.Url,
		WorkerId:   w.Id,
	}

	return p
}
