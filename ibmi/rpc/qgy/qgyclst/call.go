// Package qgyclst is the call contract of QGY/QGYCLST, Close List: it
// releases an open list's handle and server resources.
//
// Close every list on the server job that opened it, on success and on
// every error path where the session is still synchronized. If the session
// is poisoned, do not call QGYCLST: end the server job instead and IBM i
// reclaims the list.
package qgyclst

import (
	"context"
	"fmt"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/errorcode"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/openlist"
)

// Program returns the program name, QGYCLST.
func Program() rpc.ObjectName {
	n, _ := rpc.NewObjectName("QGYCLST")
	return n
}

const (
	paramErrorCode = 1
	parameterCount = 2
)

// Request describes one close.
type Request struct {
	// Library defaults to QGY (V5R4).
	Library rpc.ObjectName
	Handle  openlist.Handle
	// ErrorCode defaults to exception mode.
	ErrorCode errorcode.Parameter
	// MessageOption defaults to rpc.MessagesUpToTen.
	MessageOption rpc.MessageOption
}

// NewCall builds the QGYCLST call: handle, then error code.
func NewCall(req Request) (rpc.Call, error) {
	handle, _ := rpc.Input(req.Handle[:])
	option := req.MessageOption
	if option == 0 {
		option = rpc.MessagesUpToTen
	}
	return rpc.NewCall(qgy.LibraryOr(req.Library), Program(), []rpc.Parameter{handle, req.ErrorCode.RPC()}, option)
}

// ParseClose validates the result of a NewCall call.
func ParseClose(call rpc.Call, req Request, res *rpc.Result, jobCodec ccsid.Codec) error {
	if err := res.Validate(call); err != nil {
		return err
	}
	if call.ParameterCount() != parameterCount {
		return fmt.Errorf("qgyclst: call has %d parameters, want %d", call.ParameterCount(), parameterCount)
	}
	return req.ErrorCode.Check(call.Library(), call.Program(), res.Parameters[paramErrorCode], jobCodec)
}

// Close runs QGYCLST through caller.
func Close(ctx context.Context, caller rpc.Caller, req Request, jobCodec ccsid.Codec) error {
	call, err := NewCall(req)
	if err != nil {
		return err
	}
	res, err := caller.Call(ctx, call)
	if err != nil {
		return err
	}
	return ParseClose(call, req, res, jobCodec)
}
