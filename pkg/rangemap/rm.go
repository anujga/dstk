package rangemap

import (
	"github.com/anujga/dstk/pkg/core"
	"google.golang.org/grpc/status"
)

type RangeMap interface {
	Get(key core.KeyT) (Range, *status.Status)
	Put(rng Range) *status.Status
	Remove(rng Range) (Range, *status.Status)
}
