package simple

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/spf13/viper"
	"gopkg.in/errgo.v2/fmt/errors"
	"time"
)

type localFile struct {
	path string
}

func (l *localFile) AllParts(ctx context.Context, req *pb.AllPartsReq) (*pb.PartList, error) {
	panic("implement me")
}

func (l *localFile) MyParts(ctx context.Context, req *pb.MyPartsReq) (*pb.PartList, error) {
	panic("implement me")
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

func ClientUsingSnapshot(filename string, watch bool) (ThickClient, error) {
	r := &tc{
		t: rbt.NewWithStringComparator(),
	}
	err := core.YamlParser(filename, watch, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *tc) RefreshFile(v *viper.Viper) error {
	parts, err := parseConfig(v)
	if err != nil {
		return err
	}
	err = s.UpdateTree(parts.GetParts())
	if err != nil {
		return err
	}
	s.notifications <- time.Now()
	return nil
}
