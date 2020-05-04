package core

import (
	pb "github.com/anujga/dstk/api/protobuf-spec"
)

const (
	//note: for backward compatibility, we can only add entries in the end
	ErrUnknown               = -1
	ErrKeyNotFound           = 1
	ErrChannelFull           = iota
	ErrCantCreateSnapshotDir = iota
	ErrPeerAlreadyExists     = iota
	ErrNodeAlreadyExists     = iota

	//...
	ErrMaxErrorCode = iota
)

type Errr struct {
	pb.Ex
}

//func NewErr(id int64, msg string) *Errr {
//	return &Errr{common.Ex{Id: id, Msg: msg}}
//}
//
//func FromErr(err error) *Errr {
//	return NewErr(ErrUnknown, err.Error())
//}
//
//func FromRErr(err common.Ex) *Errr {
//	return &Errr{err}
//}

func (m *Errr) Error() string {
	return m.Msg
}

var ExOK = &pb.Ex{Id: pb.Ex_NOT_FOUND}
