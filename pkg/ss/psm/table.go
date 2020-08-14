package psm

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss/partition"
)

var TransitionTable = map[partition.State]map[partition.State]func(a partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) partition.BecomeMsg{
	partition.Init: {
		partition.CatchingUp: initToCatchingup,
		partition.Primary:    initToPrimary,
		partition.Follower:   initToFollower,
		partition.Proxy:      initToProxy,
	},
	partition.CatchingUp: {
		partition.Follower: catchingupToFollower,
	},
	partition.Primary: {
		partition.Proxy:   primaryToProxy,
		partition.Retired: primaryToRetired,
	},
	partition.Follower: {
		partition.Primary: followerToPrimary,
	},
	partition.Proxy: {
		partition.Retired: proxyToRetired,
	},
}
