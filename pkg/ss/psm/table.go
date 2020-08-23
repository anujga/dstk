package psm

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	p "github.com/anujga/dstk/pkg/ss/partition"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//todo: wrap this as a function S :: (cur, to) -> (fn)
// and add a trace. these transitions can be validated
// in background to ensure correctness.
type TransitionFn = func(a p.Actor, partIdMap map[int64]p.Actor, part *pb.Partition) p.BecomeMsg

var transitionTable = map[p.State]map[p.State]TransitionFn{
	p.Init: {
		p.CatchingUp: initToCatchingup,
		p.Primary:    initToPrimary,
		p.Follower:   initToFollower,
		p.Proxy:      initToProxy,
	},
	p.CatchingUp: {
		p.Follower: catchingupToFollower,
	},
	p.Primary: {
		p.Proxy:   primaryToProxy,
		p.Retired: primaryToRetired,
	},
	p.Follower: {
		p.Primary: followerToPrimary,
	},
	p.Proxy: {
		p.Retired: proxyToRetired,
	},
}

func GetTransition(from, to p.State) (TransitionFn, *status.Status) {
	m2, found := transitionTable[from]
	if !found {
		return nil, core.ErrInfo(codes.InvalidArgument,
			"Bad state transition state source",
			"from", from,
			"to", to)
	}

	fn, found := m2[to]
	if !found {
		return nil, core.ErrInfo(codes.InvalidArgument,
			"Bad state transition state target",
			"from", from,
			"to", to)
	}

	return fn, nil
}
