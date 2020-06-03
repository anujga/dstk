package se

import "go.uber.org/atomic"

// Interface over the various protobuf objects
// that contain partition info
type Partition interface {
	Url() string
	End() []byte
}

type RoundRobinPartition struct {
	urls []string
	end  string
	i    atomic.Uint64
	n    uint64
}

func (m *RoundRobinPartition) Name() string {
	i := m.i.Inc() % m.n
	return m.urls[i]
}
