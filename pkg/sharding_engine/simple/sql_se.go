package simple

import (
	"context"
	"fmt"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type sqlSe struct {
	db    *sqlx.DB
	clock core.DstkClock
}

func (s *sqlSe) CreatePartition(ctx context.Context, request *pb.PartitionCreateRequest) (*pb.PartitionRpcBaseResponse, error) {
	p := request.GetPartition()
	p.ModifiedOn = s.clock.Time()
	q := `INSERT INTO partition 
    		(id, modified_on, worker_id, start, "end", url, desired_state, current_state)
    	 VALUES
			(:id, :modified_on, :worker_id, :start, :end, :url, :desired_state, :current_state)
	`
	l := zap.S().With("part", p)
	result, err := s.db.NamedExec(q, *FromProto(p))
	if err != nil {
		l.Errorw("partition creation failed", "error", err, "query", q)
		return nil, status.Error(codes.Internal, "")
	}
	l.Infow("partition created", "result", result)
	return &pb.PartitionRpcBaseResponse{}, nil
}

func (s *sqlSe) GetPartitions(ctx context.Context, request *pb.PartitionGetRequest) (*pb.PartitionGetResponse, error) {
	var parts []sqlPart
	var sql string
	var err error
	var args interface{}
	// todo implement better handling of all params
	if request.GetFetchAll() {
		sql = "SELECT * FROM partition"
		err = s.db.Select(&parts, sql)
	} else {
		sql = "SELECT * FROM partition where worker_id=$1"
		args = request.GetWorkerId()
		err = s.db.Select(&parts, sql, args)
	}
	if err != nil {
		zap.S().Errorw("failed to get", "request", request, "query", sql, "args", args)
		return nil, status.Error(codes.Internal, "")
	}
	ps := toProto(parts)
	zap.S().Infow("returning parts",
		"count", len(ps.Parts),
		"client", request.WorkerId)
	return &pb.PartitionGetResponse{Partitions: ps}, nil
}

func (s *sqlSe) UpdatePartition(ctx context.Context, request *pb.PartitionUpdateRequest) (*pb.PartitionRpcBaseResponse, error) {
	updateFields := make([]string, 0)
	m := map[string]interface{}{
		"id": request.GetId(),
	}
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
	sql := fmt.Sprintf("UPDATE partition set %s where id=:id", strings.Join(updateFields, ","))
	result, err := s.db.NamedExec(sql, m)
	l := zap.S().With("req", request)
	if err != nil {
		l.Errorw("update failed", "error", err, "query", sql, "args", m)
		return nil, status.Error(codes.Internal, "")
	}
	l.Infow("update success", "result", result)
	return &pb.PartitionRpcBaseResponse{}, nil
}
