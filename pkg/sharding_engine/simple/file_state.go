package simple
//
//import (
//	"context"
//	pb "github.com/anujga/dstk/pkg/api/proto"
//	"github.com/anujga/dstk/pkg/core"
//	se "github.com/anujga/dstk/pkg/sharding_engine"
//	"github.com/spf13/viper"
//	"go.uber.org/zap"
//	"google.golang.org/grpc/codes"
//	"google.golang.org/grpc/status"
//	"sync/atomic"
//	"time"
//)
//
//type fileSe struct {
//	path  string
//	state atomic.Value
//}
//
//type seState struct {
//	parts         []*pb.Partition
//	partsByWorker map[se.WorkerId][]*pb.Partition
//	timestamp     int64
//}
//
//func (r *fileSe) AllParts(_ context.Context, _ *pb.AllPartsReq) (*pb.PartList, error) {
//	s := r.state.Load().(*seState)
//	return &pb.PartList{
//		Parts:        s.parts,
//		LastModified: s.timestamp,
//	}, nil
//}
//
//func (r *fileSe) MyParts(_ context.Context, req *pb.MyPartsReq) (*pb.PartList, error) {
//	s := r.state.Load().(*seState)
//
//	id := se.WorkerId(req.WorkerId)
//	ps, found := s.partsByWorker[id]
//	if !found {
//		return nil, core.ErrInfo(codes.InvalidArgument,
//			"Partition Not found", "workerId", string(id)).Err()
//	}
//
//	return &pb.PartList{
//		Parts:        ps,
//		LastModified: s.timestamp,
//	}, nil
//}
//
//func UsingLocalFolder(path string, watch bool) (WorkerAndClient, error) {
//	r := &fileSe{
//		path: path,
//	}
//	err := core.ParseYamlFolder(path, watch, r)
//	if err != nil {
//		return nil, err
//	}
//	return r, nil
//}
//
//func (r *fileSe) RefreshFile(v *viper.Viper) error {
//	state, err := parseConfig(v)
//	if err != nil {
//		return err.Err()
//	}
//	now := time.Now().UnixNano()
//	delay := now - state.timestamp
//	zap.S().Infow("refreshed SE",
//		"now", now,
//		"delay", delay)
//	r.state.Store(state)
//	return nil
//}
//
//type partConf struct {
//	Start, End string
//	Id         int64
//	LeaderId   int64
//}
//
//type config struct {
//	Id    int64
//	Parts []partConf
//	Url   string
//}
//
//func parseConfig(v *viper.Viper) (*seState, *status.Status) {
//	v.SetConfigName("master")
//	ws := v.GetStringSlice("workers")
//	timestamp := v.GetInt64("timestamp")
//	if timestamp <= 0 {
//		return nil, core.ErrInfo(codes.InvalidArgument,
//			"bad timestamp in master list",
//			"timestamp",
//			v.Get("timestamp"))
//	}
//
//	partMap := make(map[se.WorkerId][]*pb.Partition)
//	partList := make([]*pb.Partition, 0, 1000)
//
//	now := time.Now().UnixNano()
//	for _, w := range ws {
//		v.SetConfigName(w)
//		if err := v.ReadInConfig(); err != nil {
//			return nil, status.New(
//				codes.InvalidArgument,
//				err.Error())
//		}
//		zap.S().Infow("parsing worker", "worker", w)
//
//		var c config
//		err := v.Unmarshal(&c)
//		if err != nil {
//			return nil, status.New(
//				codes.InvalidArgument,
//				err.Error())
//		}
//
//		if c.Id == 0 {
//			return nil, core.ErrInfo(
//				codes.InvalidArgument,
//				"Bad partition id",
//				"id", v.Get("id"))
//		}
//		workerId := se.WorkerId(c.Id)
//		if _, found := partMap[workerId]; found {
//			return nil, core.ErrInfo(codes.InvalidArgument,
//				"repeated worker id in master list",
//				"workerId", workerId)
//		}
//
//		if c.Parts == nil {
//			return nil, core.ErrInfo(
//				codes.InvalidArgument,
//				"config does not contain '.parts[]'",
//				"workerId", string(workerId))
//		}
//
//		var ps = make([]*pb.Partition, 0, len(c.Parts))
//		for _, p := range c.Parts {
//			var end = []byte(p.End)
//			if len(end) == 0 {
//				end = core.MaxKey
//			}
//
//			part := pb.Partition{
//				Id:         p.Id,
//				ModifiedOn: now,
//				Active:     true,
//				Start:      []byte(p.Start),
//				End:        []byte(p.End),
//				Url:        c.Url,
//				LeaderId:   p.LeaderId,
//			}
//			ps = append(ps, &part)
//		}
//
//		partMap[workerId] = ps
//		partList = append(partList, ps...)
//	}
//
//	return &seState{
//		parts:         partList,
//		partsByWorker: partMap,
//		timestamp:     timestamp,
//	}, nil
//
//}
