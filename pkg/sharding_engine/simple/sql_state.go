package simple

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"time"
)

type sqlSe struct {
	db *sqlx.DB
}

func UsingSql(driver string, conn string) (*sqlSe, error) {
	db, err := sqlx.Connect(driver, conn)

	if err != nil {
		return nil, err
	}

	r := &sqlSe{
		db: db,
	}

	return r, nil
}

type sqlPart struct {
	Id         int64
	ModifiedOn time.Time `db:"modified_on"`
	WorkerId   int64     `db:"worker_id"`
	Start, End core.KeyT
	Url        string
}

func (s *sqlPart) toProto() *pb.Partition {
	return &pb.Partition{
		Id:         s.Id,
		ModifiedOn: s.ModifiedOn.UnixNano(),
		Active:     true,
		Start:      s.Start,
		End:        s.End,
		Url:        s.Url,
	}
}

func toProto(ps []sqlPart) *pb.PartList {
	var rs = make([]*pb.Partition, 0, len(ps))
	var lastMod = int64(0)

	for _, p := range ps {
		r := p.toProto()
		rs = append(rs, r)
		if r.ModifiedOn > lastMod {
			lastMod = r.ModifiedOn
		}
	}
	return &pb.PartList{
		Parts: rs, LastModified: lastMod,
	}
}

func (s *sqlSe) AllParts(_ context.Context, _ *pb.AllPartsReq) (*pb.PartList, error) {
	parts := []sqlPart{}
	if err := s.db.Select(&parts, "SELECT * FROM partition"); err != nil {
		return nil, err
	}

	ps := toProto(parts)
	zap.S().Infow("returning parts",
		"count", len(ps.Parts),
		"modifiedOn", ps.LastModified)
	return ps, nil
}

func (s *sqlSe) MyParts(_ context.Context, req *pb.MyPartsReq) (*pb.PartList, error) {
	parts := []sqlPart{}
	workerId := req.WorkerId
	err := s.db.Select(&parts,
		"SELECT * FROM partition where worker_id = $1",
		workerId)

	if err != nil {
		return nil, err
	}

	ps := toProto(parts)
	zap.S().Infow("returning parts",
		"workerId", workerId,
		"count", len(ps.Parts),
		"modifiedOn", ps.LastModified)
	return ps, nil
}
