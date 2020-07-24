package partition

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss/partition"
)

func initToFollower(actor partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) interface{} {
	if leader, ok := partIdMap[part.GetLeaderId()]; ok {
		return &partition.BecomeFollower{
			LeaderId: part.GetLeaderId(),
			LeaderMailbox: leader.Mailbox(),
		}
	}
	return nil
}

func initToPrimary(actor partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) interface{} {
	return &partition.BecomePrimary{}
}
