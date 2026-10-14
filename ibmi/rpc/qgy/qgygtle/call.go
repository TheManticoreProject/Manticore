// Package qgygtle is the call contract of QGY/QGYGTLE, Get List Entries:
// it reads one page of records from an open list.
//
// The handle belongs to the server job that opened the list: run QGYGTLE
// on the same session, inside the same rpc.ExclusiveCaller reservation as
// the open and close. ParsePage validates the page structurally (handle,
// record length, counts and positions); records are decoded by the
// program-qualified format that opened the list.
package qgygtle

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/errorcode"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/openlist"
)

// Program returns the program name, QGYGTLE.
func Program() rpc.ObjectName {
	n, _ := rpc.NewObjectName("QGYGTLE")
	return n
}

const (
	paramReceiver  = 0
	paramListInfo  = 3
	paramErrorCode = 6
	parameterCount = 7
)

// Request describes one page.
type Request struct {
	// Library defaults to QGY (V5R4).
	Library rpc.ObjectName
	Handle  openlist.Handle
	// RecordLength is the exact record length of the list's format.
	RecordLength uint32
	// Count is the number of records wanted (at least 1).
	Count int32
	// Start is the one-based first record wanted.
	Start int32
	// ErrorCode defaults to exception mode.
	ErrorCode errorcode.Parameter
	// MessageOption defaults to rpc.MessagesUpToTen.
	MessageOption rpc.MessageOption
}

// receiverLength is Count*RecordLength, checked to fit BINARY(4).
func (r Request) receiverLength() (uint32, error) {
	if r.RecordLength == 0 {
		return 0, errors.New("qgygtle: zero record length")
	}
	if r.Count < 1 {
		return 0, fmt.Errorf("qgygtle: record count %d must be at least 1", r.Count)
	}
	if r.Start < 1 {
		return 0, fmt.Errorf("qgygtle: starting record %d must be at least 1", r.Start)
	}
	n := uint64(r.Count) * uint64(r.RecordLength)
	if n > math.MaxInt32 {
		return 0, fmt.Errorf("qgygtle: receiver of %d bytes exceeds the BINARY(4) maximum", n)
	}
	return uint32(n), nil
}

func input(b []byte) rpc.Parameter {
	p, _ := rpc.Input(b)
	return p
}

// NewCall builds the QGYGTLE call. Parameters, in order: receiver (output
// Count*RecordLength), receiver length, handle, list information (output
// 80), records to return, starting record, error code.
func NewCall(req Request) (rpc.Call, error) {
	n, err := req.receiverLength()
	if err != nil {
		return rpc.Call{}, err
	}
	receiver, err := rpc.Output(n)
	if err != nil {
		return rpc.Call{}, err
	}
	listInfo, _ := rpc.Output(openlist.Length)
	option := req.MessageOption
	if option == 0 {
		option = rpc.MessagesUpToTen
	}
	return rpc.NewCall(qgy.LibraryOr(req.Library), Program(), []rpc.Parameter{
		receiver,
		input(qgy.Binary4(int32(n))),
		input(req.Handle[:]),
		listInfo,
		input(qgy.Binary4(req.Count)),
		input(qgy.Binary4(req.Start)),
		req.ErrorCode.RPC(),
	}, option)
}

// Page is one validated page.
type Page struct {
	Info openlist.Info
	// Records holds exactly Info.RecordsReturned records, each
	// RecordLength bytes; decode them with the opening program's format.
	Records []byte
}

// ErrInvalidPage is matched when a page's metadata is inconsistent with the
// request or with itself.
var ErrInvalidPage = errors.New("qgygtle: invalid page")

// ParsePage validates the result of a NewCall call: same handle, the
// request's record length, no more records than requested, the requested
// first record when any were returned, and enough returned bytes. It does
// not judge list status; the pager decides what a status means.
func ParsePage(call rpc.Call, req Request, res *rpc.Result, jobCodec ccsid.Codec) (*Page, error) {
	if err := res.Validate(call); err != nil {
		return nil, err
	}
	if call.ParameterCount() != parameterCount {
		return nil, fmt.Errorf("qgygtle: call has %d parameters, want %d", call.ParameterCount(), parameterCount)
	}
	if err := req.ErrorCode.Check(call.Library(), call.Program(), res.Parameters[paramErrorCode], jobCodec); err != nil {
		return nil, err
	}
	raw, ok := res.Output(paramListInfo)
	if !ok {
		return nil, fmt.Errorf("%w: list information was not returned", ErrInvalidPage)
	}
	info, err := openlist.Unmarshal(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPage, err)
	}
	receiver, _ := res.Output(paramReceiver)
	need := uint64(info.RecordsReturned) * uint64(req.RecordLength)
	switch {
	case info.Handle != req.Handle:
		return nil, fmt.Errorf("%w: the list handle changed", ErrInvalidPage)
	case uint32(info.RecordLength) != req.RecordLength:
		return nil, fmt.Errorf("%w: record length %d, want %d", ErrInvalidPage, info.RecordLength, req.RecordLength)
	case info.RecordsReturned > req.Count:
		return nil, fmt.Errorf("%w: %d records returned for %d requested", ErrInvalidPage, info.RecordsReturned, req.Count)
	case info.RecordsReturned > 0 && info.FirstRecord != req.Start:
		return nil, fmt.Errorf("%w: first record %d, requested %d", ErrInvalidPage, info.FirstRecord, req.Start)
	case need > uint64(len(receiver)):
		return nil, fmt.Errorf("%w: %d records need %d bytes, receiver has %d", ErrInvalidPage, info.RecordsReturned, need, len(receiver))
	case need > uint64(info.InfoLength):
		return nil, fmt.Errorf("%w: %d records need %d bytes, server returned %d", ErrInvalidPage, info.RecordsReturned, need, info.InfoLength)
	}
	return &Page{Info: info, Records: receiver[:need:need]}, nil
}

// Get runs QGYGTLE through caller.
func Get(ctx context.Context, caller rpc.Caller, req Request, jobCodec ccsid.Codec) (*Page, error) {
	call, err := NewCall(req)
	if err != nil {
		return nil, err
	}
	res, err := caller.Call(ctx, call)
	if err != nil {
		return nil, err
	}
	return ParsePage(call, req, res, jobCodec)
}
