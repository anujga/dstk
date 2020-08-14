package psm

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss/partition"
)

func followerToPrimary(actor partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) partition.BecomeMsg {
	return &partition.BecomeMsgImpl{TargetState: partition.Primary}
}
