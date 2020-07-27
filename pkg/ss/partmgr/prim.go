package partition

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss/partition"
)

func primaryToProxy(actor partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) interface{} {
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

func primaryToRetired(actor partition.Actor, partIdMap map[int64]partition.Actor, part *pb.Partition) interface{} {
	return &partition.Retire{}
}
