package simple

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"sync/atomic"
	"time"
)

type fileSe struct {
	path  string
	state atomic.Value
}

type seState struct {
	parts         []*pb.Partition
	partsByWorker map[se.WorkerId][]*pb.Partition
	timestamp     int64
}

func (r *fileSe) AllParts(_ context.Context, _ *pb.AllPartsReq) (*pb.PartList, error) {
	s := r.state.Load().(*seState)
	return &pb.PartList{
		Parts:        s.parts,
		LastModified: s.timestamp,
	}, nil
}

func (r *fileSe) MyParts(_ context.Context, req *pb.MyPartsReq) (*pb.PartList, error) {
	s := r.state.Load().(seState)

	id := se.WorkerId(req.WorkerId)
	ps, found := s.partsByWorker[id]
	if !found {
		return nil, core.ErrInfo(codes.InvalidArgument,
			"Partition Not found", "workerId", string(id)).Err()
	}

	return &pb.PartList{
		Parts:        ps,
		LastModified: s.timestamp,
	}, nil
}

func UsingLocalFolder(path string, watch bool) (pb.SeWorkerApiServer, error) {
	r := &fileSe{
		path: path,
	}
	err := core.ParseYamlFolder(path, watch, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *fileSe) RefreshFile(v *viper.Viper) error {
	state, err := parseConfig(v)
	if err != nil {
		return err.Err()
	}
	now := time.Now().UnixNano()
	delay := now - state.timestamp
	zap.S().Infow("refreshed SE",
		"now", now,
		"delay", delay)
	r.state.Store(state)
	return nil
}

func parseConfig(v *viper.Viper) (*seState, *status.Status) {
	v.SetConfigName("master")
	ws := v.GetStringSlice("workers")
	timestamp := v.GetInt64("timestamp")
	if timestamp <= 0 {
		return nil, core.ErrInfo(codes.InvalidArgument,
			"bad timestamp in master list",
			"timestamp",
			v.Get("timestamp"))
	}

	partMap := make(map[se.WorkerId][]*pb.Partition)
	partList := make([]*pb.Partition, 1000)

	for _, w := range ws {
		v.SetConfigName(w)
		zap.S().Infow("parsing worker", "worker", w)

		id, err := strconv.ParseInt(w, 10, 64)
		if err != nil {
			return nil, status.New(codes.InvalidArgument, err.Error())
		}
		workerId := se.WorkerId(id)
		if _, found := partMap[workerId]; found {
			return nil, core.ErrInfo(codes.InvalidArgument,
				"repeated worker id in master list",
				"workerId", workerId)
		}

		parts := v.Get("parts")
		if parts == nil {
			return nil, core.ErrInfo(
				codes.InvalidArgument,
				"config does not contain '.parts[]'",
				"workerId", string(workerId))
		}

		ps := &pb.Partitions{}
		err = core.Obj2Proto(parts, ps)
		if err != nil {
			return nil, status.New(codes.InvalidArgument, err.Error())
		}

		partMap[workerId] = ps.Parts
		partList = append(partList, ps.Parts...)
	}

	return &seState{
		parts:         partList,
		partsByWorker: partMap,
		timestamp:     timestamp,
	}, nil

}
