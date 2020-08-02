package simple

import (
	"context"
	"fmt"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"strings"
	"time"
)

type sqlSe struct {
	db *sqlx.DB
}

func (s *sqlSe) CreatePartition(ctx context.Context, request *pb.PartitionCreateRequest) (*pb.PartitionRpcBaseResponse, error) {
	p := request.GetPartition()
	p.ModifiedOn = time.Now().Unix()
	q := `INSERT INTO partition 
    		(id, modified_on, worker_id, start, "end", url, desired_state, current_state)
    	 VALUES
			(:id, :modified_on, :worker_id, :start, :end, :url, :desired_state, :current_state)
	`
	l := zap.L().With(zap.Any("part", p))
	if result, err := s.db.NamedExec(q, *FromProto(p)); err == nil {
		l.Info("partition created", zap.Any("result", result))
		return &pb.PartitionRpcBaseResponse{}, nil
	} else {
		l.Error("partition creation failed", zap.Error(err))
		return nil, err
	}
}

func (s *sqlSe) GetPartitions(ctx context.Context, request *pb.PartitionsGetRequest) (*pb.PartitionGetResponse, error) {
	var parts []sqlPart
	var sql string
	// todo implement better handling of all params
	if request.GetWorkerId() == 0 {
		sql = "SELECT * FROM partition"
	} else {
		sql = fmt.Sprintf("SELECT * FROM partition where worker_id=%d", request.GetWorkerId())
	}
	if err := s.db.Select(&parts, sql); err != nil {
		return nil, err
	}
	ps := toProto(parts)
	zap.S().Infow("returning parts", "count", len(ps.Parts))
	return &pb.PartitionGetResponse{Partitions: ps}, nil
}

func (s *sqlSe) UpdatePartition(ctx context.Context, request *pb.PartitionUpdateRequest) (*pb.PartitionRpcBaseResponse, error) {
	updateFields := make([]string, 0)
	m := make(map[string]interface{})
	if request.GetCurrentState() != "" {
		f := "current_state"
		updateFields = append(updateFields, fmt.Sprintf("%s=:%s", f, f))
		m[f] = request.GetCurrentState()
	}
	if request.GetDesiredState() != "" {
		f := "desired_state"
		updateFields = append(updateFields, fmt.Sprintf("%s=:%s", f, f))
		m[f] = request.GetDesiredState()
	}
	// todo make use of etag
	sql := fmt.Sprintf("UPDATE partition set %s where id=%d", strings.Join(updateFields, ","), request.GetId())
	result, err := s.db.NamedExec(sql, m)
	l := zap.L().With(zap.Any("req", request))
	if err != nil {
		l.Error("update failed", zap.Error(err))
		return nil, err
	}
	l.Info("update success", zap.Any("result", result))
	return &pb.PartitionRpcBaseResponse{}, nil
}
