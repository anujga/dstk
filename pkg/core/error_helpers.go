package core

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	details "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	//note: for backward compatibility, we can only add entries in the end
	ErrUnknown               = -1
	ErrKeyNotFound           = 1
	ErrInvalidPartition      = 2
	ErrChannelFull           = iota
	ErrCantCreateSnapshotDir = iota
	ErrPeerAlreadyExists     = iota
	ErrNodeAlreadyExists     = iota

	//...
	ErrMaxErrorCode = iota
)

type Errr struct {
	*dstk.Ex
}

func NewErr(id dstk.Ex_ExCode, msg string) *Errr {
	return &Errr{&dstk.Ex{Id: id, Msg: msg}}
}

func WrapEx(err *dstk.Ex) *Errr {
	return &Errr{err}
}

func (m *Errr) Error() string {
	return m.Msg
}

var ExOK = &dstk.Ex{Id: dstk.Ex_SUCCESS}

func ErrInfo(c codes.Code, msg string, k string, v string) *status.Status {
	ex := status.New(c, msg)

	ex2, err := ex.WithDetails(
		&details.ErrorInfo{
			Metadata: map[string]string{k: v},
		})

	if err != nil {
		return ex
	} else {
		return ex2
	}
}
