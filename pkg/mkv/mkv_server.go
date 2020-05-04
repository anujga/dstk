package mkv

import (
	"context"
	pb "github.com/anujga/dstk/api/protobuf-spec"
	"github.com/anujga/dstk/pkg/core"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"io/ioutil"
	"strings"
	"sync"
)

type MapEntry struct {
	payload     []byte
	partitionId int64
}

type mkvServer struct {
	data       map[string]MapEntry
	partitions map[int64]string
	mu         sync.Mutex

	log  *zap.Logger
	slog *zap.SugaredLogger
}

func MakeServer(port int32) (pb.MkvServer, error) {
	s := mkvServer{
		data:       make(map[string]MapEntry),
		partitions: make(map[int64]string),
	}
	var err error
	if s.log, err = zap.NewProduction(); err != nil {
		return nil, err
	}

	s.slog = s.log.Sugar()
	return &s, nil
}

func (s *mkvServer) AddPart(ctx context.Context, args *pb.AddParReq) (*pb.Ex, error) {
	uri := args.GetUri()
	if !strings.HasPrefix(uri, "file://") {
		s.slog.Errorw("Unkown file prefix",
			"allowed", "file://",
			"found", uri,
		)
		return &pb.Ex{Id: pb.Ex_NOT_IMPLEMENTED}, nil
	}

	filename := strings.TrimPrefix(uri, "file://")
	s.slog.Infow("adding partition",
		// Structured context as loosely typed key-value pairs.
		"filename", filename,
	)

	fin, err := ioutil.ReadFile(filename)
	if err != nil {
		s.slog.Errorw("Could not open patition file",
			"uri", uri,
			"err", err)
		return &pb.Ex{Id: pb.Ex_INVALID_ARGUMENT}, err
	}

	p := &pb.Partition{}
	if err := proto.Unmarshal(fin, p); err != nil {
		s.slog.Errorw("Failed to parse partition file:",
			"uri", uri,
			"err", err)
	}

	id := p.GetId()
	s.mu.Lock()
	if _, ok := s.partitions[id]; ok {
		s.slog.Errorw("Partition already exists")
	}

	i := 0
	var e *pb.Partition_Entry
	for i, e = range p.Entries {
		s.mu.Lock()
		s.data[string(e.Key)] = MapEntry{
			partitionId: p.GetId(),
			payload:     e.Value,
		}
		s.mu.Unlock()
	}
	s.slog.Infow("file read successfully", "uri", uri, "count", i)
	return core.ExOK, nil
}

func (s *mkvServer) Get(ctx context.Context, args *pb.GetReq) (*pb.GetRes, error) {
	k := string(args.GetKey())
	p := args.GetPartitionId()
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.partitions[p]; !ok {
		return &pb.GetRes{Ex: &pb.Ex{Id: pb.Ex_BAD_PARTITION}}, nil
	}

	val, ok := s.data[k]
	if !ok {
		return &pb.GetRes{Ex: &pb.Ex{Id: pb.Ex_NOT_FOUND}}, nil
	}

	return &pb.GetRes{
		Ex:          core.ExOK,
		PartitionId: val.partitionId,
		Payload:     val.payload,
	}, nil
}
