package simple

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
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

func InitDb(driver string, conn string, parts []*pb.Partition) error {
	db, err := sqlx.Connect(driver, conn)
	if err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	err = add(tx, parts)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			panic(err2)
		}
		return err
	}
	err = tx.Commit()
	return err
}

func add(tx *sqlx.Tx, parts []*pb.Partition) error {

	_, err := tx.Exec("delete from partition")
	if err != nil {
		return err
	}

	q := `INSERT INTO partition 
    		(id, modified_on, worker_id, start, "end", url)
    	 VALUES
			(:id, :modified_on, :worker_id, :start, :end, :url)
	`

	for _, p := range parts {
		p2 := FromProto(p)
		res, err := tx.NamedExec(q, p2)
		if err != nil {
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if rows != 1 {
			return core.ErrInfo(codes.InvalidArgument,
				"Failed to insert row",
				"part", p).Err()
		}
	}
	return nil
}

type sqlPart struct {
	Id           int64
	ModifiedOn   time.Time `db:"modified_on"`
	WorkerId     int64     `db:"worker_id"`
	Start, End   core.KeyT
	Url          string
	LeaderId     int64  `db:"leader_id"`
	ProxyTo      int64  `db:"proxy_to"`
	DesiredState string `db:"desired_state"`
}

func FromProto(p *pb.Partition) *sqlPart {
	return &sqlPart{
		Id:           p.GetId(),
		ModifiedOn:   time.Unix(0, p.GetModifiedOn()),
		WorkerId:     p.GetWorkerId(),
		Start:        p.GetStart(),
		End:          p.GetEnd(),
		Url:          p.GetUrl(),
		LeaderId:     p.GetLeaderId(),
		DesiredState: p.GetDesiredState(),
		ProxyTo:      p.GetProxyTo(),
	}
}

func (s *sqlPart) toProto() *pb.Partition {
	return &pb.Partition{
		Id:           s.Id,
		ModifiedOn:   s.ModifiedOn.UnixNano(),
		Active:       true,
		Start:        s.Start,
		End:          s.End,
		Url:          s.Url,
		LeaderId:     s.LeaderId,
		DesiredState: s.DesiredState,
		ProxyTo:      s.ProxyTo,
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
	var parts []sqlPart
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
	var parts []sqlPart
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
