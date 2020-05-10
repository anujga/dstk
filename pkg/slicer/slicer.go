package slicer

import (
	"github.com/anujga/dstk/pkg/core"
)

type SliceRdr interface {
	Get(key []byte) (Partition, error)
}

type slicerCli struct {
	cli      SliceRdr
	connPool core.ConnPool
}
