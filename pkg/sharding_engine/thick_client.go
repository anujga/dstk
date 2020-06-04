package se

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/errgo.v2/fmt/errors"
	"sync"
	"time"
)

//todo: implement using immutable tree and get rid of mutex
type snapshotFile struct {
	t             *rbt.Tree
	mu            sync.Mutex
	lastPart      *pb.Partition
	notifications chan interface{}
}

func (s *snapshotFile) Notifications() <-chan interface{} {
	return s.notifications
}

func ClientUsingSnapshot(filename string, watch bool) (ThickClient, error) {
	r := &snapshotFile{
		t: rbt.NewWithStringComparator(),
	}
	err := core.YamlParser(filename, watch, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *snapshotFile) RefreshFile(v *viper.Viper) error {
	parts, err := parseConfig(v)
	if err != nil {
		return err
	}
	err = s.updateTree(parts.GetParts())
	if err != nil {
		return err
	}
	s.notifications <- time.Now()
	return nil
}

func (s *snapshotFile) Parts() []*pb.Partition {
	var iface []*pb.Partition
	s.mu.Lock()
	{
		ps := s.t.Values()
		iface = make([]*pb.Partition, len(ps))
		for i := range ps {
			iface[i] = ps[i].(*pb.Partition)
		}
	}
	s.mu.Unlock()
	return iface
}

func parseConfig(v *viper.Viper) (*pb.Partitions, error) {
	parts := v.Get("parts")
	if parts == nil {
		return nil, errors.Newf("config does not contain '.parts[]'")
	}

	var ps pb.Partitions
	err := core.Obj2Proto(parts, &ps)
	if err != nil {
		return nil, err
	}
	return &ps, nil
}

func (s *snapshotFile) updateTree(parts []*pb.Partition) error {
	var lastPart *pb.Partition
	t := rbt.NewWithStringComparator()

	for _, p := range parts {
		k := string(p.GetEnd())
		if k == "" {
			if lastPart != nil {
				return errors.Newf(
					"Only 1 msg can have end = empty. found %s, %s",
					s.lastPart.Id, p.Id)
			}
			lastPart = p
		} else {
			t.Put(k, p)
		}
	}

	if lastPart == nil {
		return errors.Newf("End partition not found")
	}

	zap.S().Infow("Partitions found", "count", len(parts))

	s.mu.Lock()
	s.t = t
	s.lastPart = lastPart
	s.mu.Unlock()
	return nil
}

func (s *snapshotFile) Get(ctx context.Context, key []byte) (*pb.Partition, error) {
	k := string(key)
	s.mu.Lock()
	v, found := s.t.Ceiling(k)
	s.mu.Unlock()
	if !found {
		return s.lastPart, nil
	}

	return v.Value.(*pb.Partition), nil
}
