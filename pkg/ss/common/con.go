package common

import dstk "github.com/anujga/dstk/pkg/api/proto"

type Consumer interface {
	Process(msg Msg) (interface{}, error)
	GetSnapshot() AppState
	ApplySnapshot(as AppState) error
}

type ConsumerFactory interface {
	Make(p *dstk.Partition) (Consumer, int, error)
}
