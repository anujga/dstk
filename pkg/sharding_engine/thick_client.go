package se

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/errgo.v2/fmt/errors"
	"sync"
)

//todo: implement using immutable tree and get rid of mutex
type staticClient struct {
	t        *rbt.Tree
	mu       sync.Mutex
	lastPart *pb.Partition
}

func (s *staticClient) Update(parts []*pb.Partition) error {
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

func StaticFile(filename string, watch bool) (ThickClient, error) {
	v, err := core.ParseYaml(filename)
	if err != nil {
		return nil, err
	}

	r := &staticClient{
		t: rbt.NewWithStringComparator(),
	}

	err = refreshFile(v, r)
	if err != nil {
		return nil, err
	}

	if watch {
		v.WatchConfig()
		v.OnConfigChange(func(in fsnotify.Event) {
			filename := v.ConfigFileUsed()
			zap.S().Infow("config file modified", "filename", filename)
			if (in.Op & (fsnotify.Write | fsnotify.Create)) == 0 {
				return
			}
			zap.S().Infow("config parsing file", "filename", filename)

			err := refreshFile(v, r)
			if err != nil {
				zap.S().Errorw("failed to refresh config",
					"filename", filename,
					"error", err)
			}
		})
	}

	return r, nil
}

func getParts(v *viper.Viper) (*pb.Partitions, error) {
	parts := v.Get("parts")
	if parts == nil {
		return nil, errors.Newf("config does not contain '.parts[]'")
	}

	var m pb.Partitions
	err := core.Obj2Proto(parts, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func refreshFile(v *viper.Viper, r *staticClient) error {
	parts, err := getParts(v)
	if err != nil {
		return err
	}
	err = r.Update(parts.GetParts())
	if err != nil {
		return err
	}
	return nil
}

func (s *staticClient) Get(ctx context.Context, key []byte) (*pb.Partition, error) {
	k := string(key)
	s.mu.Lock()
	v, found := s.t.Ceiling(k)
	s.mu.Unlock()
	if !found {
		return s.lastPart, nil
	}

	return v.Value.(*pb.Partition), nil
}
