package partition

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss/partition"
)

func primaryToProxy(actor partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) interface{} {
	if a, ok := partIdMap[part.GetLeaderId()]; ok {
		bp := &partition.BecomeProxy{
			ProxyToId: a.Id(),
			ProxyTo:   a.Mailbox(),
		}
		return bp
	}
	return nil
}
