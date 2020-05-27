package sharder

import (
	"context"
	dstk "github.com/anujga/dstk/pkg/api/proto"
)

type Client interface {
	FindPartition(_ context.Context, in *dstk.Find_Req) (*dstk.Find_Res, error)
}
