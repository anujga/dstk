package simple

import (
	"database/sql"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"time"
)

func UsingSql(driver string, conn string) (*sqlSe, error) {
	db, err := sqlx.Connect(driver, conn)

	if err != nil {
		return nil, err
	}

	r := &sqlSe{
		db: db,
		clock: &core.RealClock{},
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
    		(id, modified_on, worker_id, start, "end", url, desired_state)
    	 VALUES
			(:id, :modified_on, :worker_id, :start, :end, :url, :desired_state)
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
	ModifiedOn   time.Time     `db:"modified_on"`
	WorkerId     int64         `db:"worker_id"`
	Start        core.KeyT     `db:"start"`
	End          core.KeyT     `db:"end"`
	Url          string        `db:"url"`
	LeaderId     sql.NullInt64 `db:"leader_id"`
	ProxyTo      pq.Int64Array `db:"proxy_to"`
	DesiredState string        `db:"desired_state"`
	CurrentState string        `db:"current_state"`
}

func FromProto(p *pb.Partition) *sqlPart {
	sp := &sqlPart{
		Id:           p.GetId(),
		ModifiedOn:   time.Unix(0, p.GetModifiedOn()),
		WorkerId:     p.GetWorkerId(),
		Start:        p.GetStart(),
		End:          p.GetEnd(),
		Url:          p.GetUrl(),
		DesiredState: p.GetDesiredState(),
		ProxyTo:      p.GetProxyTo(),
		CurrentState: p.GetCurrentState(),
	}
	if p.GetLeaderId() != 0 {
		sp.LeaderId = sql.NullInt64{
			Int64: p.GetLeaderId(),
			Valid: true,
		}
	}
	return sp
}

func (s *sqlPart) toProto() *pb.Partition {
	p := &pb.Partition{
		Id:           s.Id,
		ModifiedOn:   s.ModifiedOn.UnixNano(),
		Active:       true,
		Start:        s.Start,
		End:          s.End,
		Url:          s.Url,
		DesiredState: s.DesiredState,
		ProxyTo:      s.ProxyTo,
		CurrentState: s.CurrentState,
	}
	if s.LeaderId.Valid {
		p.LeaderId = s.LeaderId.Int64
	}
	return p
}

func toProto(ps []sqlPart) *pb.Partitions {
	var rs = make([]*pb.Partition, 0, len(ps))
	var lastMod = int64(0)

	for _, p := range ps {
		r := p.toProto()
		rs = append(rs, r)
		if r.ModifiedOn > lastMod {
			lastMod = r.ModifiedOn
		}
	}
	return &pb.Partitions{Parts: rs}
}
