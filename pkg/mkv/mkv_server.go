package mkv

import (
	"context"
	"fmt"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"io/ioutil"
	"net"
	"os"
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

func MakeServer() (pb.MkvServer, error) {
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

func (s *mkvServer) AddPart(ctx context.Context, args *pb.AddParReq) error {
	uri := args.GetUri()
	if !strings.HasPrefix(uri, "file://") {
		s.slog.Errorw("Unkown file prefix",
			"allowed", "file://",
			"found", uri,
		)
		return status.Newf(codes.Unimplemented, "Url should have a `file://` prefix. found %s", uri).Err()
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
		return err
	}

	p := &pb.MkvPartition{}
	if err := proto.Unmarshal(fin, p); err != nil {
		s.slog.Errorw("Failed to parse partition file:",
			"uri", uri,
			"err", err)

		return err
	}

	id := p.GetId()
	err = s.insertPartition(id, uri)
	if err != nil {
		return err
	}

	i := 0
	var e *pb.MkvPartition_Entry
	for i, e = range p.Entries {
		s.mu.Lock()
		s.data[string(e.Key)] = MapEntry{
			partitionId: p.GetId(),
			payload:     e.Value,
		}
		s.mu.Unlock()
	}

	s.slog.Infow("file read successfully", "uri", uri, "count", i+1)
	return nil
}

func (s *mkvServer) insertPartition(id int64, uri string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if oldPart, ok := s.partitions[id]; ok {
		s.slog.Errorw("Partition already exists")
		return core.ErrInfo(
			codes.InvalidArgument,
			"Partition already exists",
			"old", oldPart).Err()
	}

	//note: partition is already added assuming there cannot be any failure
	//subsequently
	s.partitions[id] = uri

	return nil
}

func (s *mkvServer) Get(ctx context.Context, args *pb.GetReq) (*pb.GetRes, error) {
	k := string(args.GetKey())
	p := args.GetPartitionId()
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.partitions[p]; !ok {
		return nil, core.ErrInfo(codes.Unavailable, "Bad Partition",
			"resolved", string(p)).Err()
	}

	val, ok := s.data[k]
	if !ok {
		return nil, status.New(codes.NotFound, "Not found").Err()
	}

	return &pb.GetRes{
		PartitionId: val.partitionId,
		Payload:     val.payload,
	}, nil
}

func StartServer(port int32, listener *bufconn.Listener) (*grpc.Server, func()) {
	log, err := zap.NewProduction()
	if err != nil {
		println("Failed to open logger %s", err)
		os.Exit(-1)
	}
	slog := log.Sugar()

	var lis net.Listener
	if port == 0 {
		lis = listener
	} else {
		lis, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			slog.Fatalw("failed to listen",
				"port", port,
				"err", err)
		}
	}
	grpcServer := grpc.NewServer()
	s, err := MakeServer()
	if err != nil {
		slog.Fatalw("failed to initialize server object",
			"port", port,
			"err", err)
	}

	pb.RegisterMkvServer(grpcServer, s)

	return grpcServer, func() {
		if err = grpcServer.Serve(lis); err != nil {
			slog.Fatalw("failed to start server",
				"port", port,
				"err", err)
			//todo: dont crash the process, return a promise or channel
			//or create a small interface for shutdown, shutdownNow, didStart ...
			os.Exit(-2)
		}
	}
}
