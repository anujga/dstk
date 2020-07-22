package core

import (
	"fmt"
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

type MultiErr struct {
	errs []error
}

func Errs(errs ...error) error {
	var es []error
	for _, e := range errs {
		if e != nil {
			es = append(es, e)
		}
	}
	if es == nil {
		return nil
	}

	return &MultiErr{errs: es}
}

func (e *MultiErr) Error() string {
	return fmt.Sprintf("%+q", e.errs)
}

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

func values2Map(keyValues ...interface{}) map[string]string {
	n := len(keyValues)
	kv := make(map[string]string, n/2)
	for i := 0; i < n; i += 2 {
		var v = ""
		if i+1 < n {
			v = fmt.Sprintf("%#v", keyValues[i+1])
		}
		k := fmt.Sprintf("%#v", keyValues[i])
		kv[k] = v
	}
	return kv
}

func ErrInfo(c codes.Code, msg string, keyValues ...interface{}) *status.Status {
	ex := status.New(c, msg)

	ex2, err := ex.WithDetails(
		&details.ErrorInfo{
			Metadata: values2Map(keyValues),
		})

	if err != nil {
		return ex
	} else {
		return ex2
	}
}
