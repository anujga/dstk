package rangemap

import (
	"github.com/anujga/dstk/pkg/core"
	"google.golang.org/grpc/status"
	"io"
)

type RangeMap interface {
	io.Closer
	Get(key core.KeyT) (Range, bool, *status.Status)
	Put(rng Range) *status.Status
	Remove(rng Range) (Range, *status.Status)
}
