package ss

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/rangemap"
	"go.uber.org/zap"
)

type appState struct {
	s interface{}
}

func (a *appState) ResponseChannel() chan interface{} {
	return nil
}

func (a *appState) State() interface{} {
	return a.s
}

type state struct {
	m            *rangemap.RangeMap
	lastModified int64
	logger       *zap.SugaredLogger
}

// path=control
func (s *state) add(p *pb.Partition, consumer ConsumerFactory, caughtUpListener func(*PartRange)) (*PartRange, error) {
	var err error
	s.logger.Infow("AddPartition Start", "part", p)
	defer s.logger.Infow("AddPartition Status", "part", p, "err", err)
	c, maxOutstanding, err := consumer.Make(p)
	if err != nil {
		return nil, err
	}
	part := NewPartRange(p, c, maxOutstanding, caughtUpListener)
	if err = s.m.Put(part); err != nil {
		return nil, err
	}
	part.Run()
	return part, nil
}

func (s *state) addPart(pa *PartRange) error {
	return s.m.Put(pa)
}

func (s *state) removePart(pa *PartRange) (*PartRange, error) {
	if pr, err := s.m.Remove(pa); err == nil {
		return pr.(*PartRange), err
	} else {
		return nil, err
	}
}
