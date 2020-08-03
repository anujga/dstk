package partition

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/ss/common"
	"github.com/anujga/dstk/pkg/ss/partition"
	"github.com/google/btree"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

type Manager interface {
	Find(key core.KeyT) (partition.Actor, error)
	Reset(plist *pb.Partitions) error
}

// todo: this is the 4th implementation of the range map.
// need to define a proper data structure that can be reused
type managerImpl struct {
	store           *actorStore
	consumerFactory common.ConsumerFactory
	slog            *zap.SugaredLogger
}

func (pm *managerImpl) Find(key core.KeyT) (partition.Actor, error) {
	return pm.store.find(key)
}

func (pm *managerImpl) Reset(plist *pb.Partitions) error {
	return resetParts(plist, pm, pm.slog.Desugar())
}

//todo: ensure there is at least 1 partition during construction
func NewManager(factory common.ConsumerFactory) (Manager, *status.Status) {
	return &managerImpl{
		consumerFactory: factory,
		slog:            zap.S(),
		store: &actorStore{
			partRoot:     btree.New(16),
			partIdMap:    make(map[int64]partition.Actor),
			lastModified: 0,
		},
	}, nil
}
