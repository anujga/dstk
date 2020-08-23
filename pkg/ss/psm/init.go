package psm

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss/partition"
)

func initToProxy(actor partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) partition.BecomeMsg {
	pt := make([]partition.Actor, 0)
	for _, pId := range part.GetProxyTo() {
		if a, ok := partIdMap[pId]; ok {
			pt = append(pt, a)
		} else {
			//todo handle
		}
	}
	return &partition.BecomeProxy{ProxyTo: pt}
}

func initToCatchingup(actor partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) partition.BecomeMsg {
	if leader, ok := partIdMap[part.GetLeaderId()]; ok {
		return &partition.BecomeCatchingUpActor{
			LeaderId:      part.GetLeaderId(),
			LeaderMailbox: leader.Mailbox(),
		}
	}
	return nil
}

func initToPrimary(actor partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) partition.BecomeMsg {
	return &partition.BecomeMsgImpl{TargetState: partition.Primary}
}

func initToFollower(actor partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) partition.BecomeMsg {
	return &partition.BecomeMsgImpl{TargetState: partition.Follower}
}
