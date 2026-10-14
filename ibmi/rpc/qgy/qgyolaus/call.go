// Package qgyolaus is the call contract of QGY/QGYOLAUS, Open List of
// Authorized Users.
//
// NewCall builds the neutral rpc.Call; ParseOpen validates its result; Open
// does both through an rpc.Caller. The list is built synchronously (records
// to return = -1) with a zero-length receiver, as the V5R4-era JTOpen 4.8
// client does, so the open call returns only list information and a handle;
// records are read with qgygtle and the handle released with qgyclst on the
// same server job.
//
// IBM i returns only profiles the caller has *READ authority to, and
// QGYOLAUS is not thread-safe: run open, get and close inside one
// rpc.ExclusiveCaller reservation.
package qgyolaus

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/errorcode"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/openlist"
)

// Program returns the program name, QGYOLAUS.
func Program() rpc.ObjectName {
	n, _ := rpc.NewObjectName("QGYOLAUS")
	return n
}

// Parameter positions in the call.
const (
	paramListInfo  = 2
	paramErrorCode = 7
	parameterCount = 9
)

// synchronous asks the server to build the whole list before returning.
const synchronous int32 = -1

// Selection is the selection-criteria parameter.
type Selection uint8

const (
	// SelectAll lists users and groups (*ALL).
	SelectAll Selection = iota
	// SelectUsers lists profiles that are not groups (*USER).
	SelectUsers
	// SelectGroups lists group profiles (*GROUP).
	SelectGroups
	// SelectMembers lists the members of Request.Group (*MEMBER).
	SelectMembers
)

func (s Selection) value() (string, error) {
	switch s {
	case SelectAll:
		return "*ALL", nil
	case SelectUsers:
		return "*USER", nil
	case SelectGroups:
		return "*GROUP", nil
	case SelectMembers:
		return "*MEMBER", nil
	}
	return "", fmt.Errorf("qgyolaus: undefined selection %d", s)
}

// Request describes one QGYOLAUS open.
type Request struct {
	// Library defaults to QGY (V5R4); current releases also accept QSYS.
	Library   rpc.ObjectName
	Selection Selection
	// Group is required for SelectMembers and must be empty otherwise.
	Group string
	// Profile is a profile name or IBM generic name (prefix ending in *);
	// empty means *ALL.
	Profile string
	// ErrorCode defaults to exception mode.
	ErrorCode errorcode.Parameter
	// MessageOption defaults to rpc.MessagesUpToTen.
	MessageOption rpc.MessageOption
}

var errZeroFormat = errors.New("qgyolaus: zero Format; use AUTU0100, AUTU0150, AUTU0200 or AUTU0250")

func errRecords(count int, length uint32, have int) error {
	return fmt.Errorf("qgyolaus: %d records of %d bytes do not fit a %d-byte receiver", count, length, have)
}

// validateProfile accepts *ALL, a system name, or a generic name (a valid
// name prefix followed by one trailing *).
func validateProfile(p string) (string, error) {
	if p == "" || strings.EqualFold(p, "*ALL") {
		return "*ALL", nil
	}
	prefix, generic := strings.CutSuffix(p, "*")
	if generic && len(p) > rpc.MaxObjectNameLength {
		return "", fmt.Errorf("qgyolaus: generic profile name %q longer than %d characters", p, rpc.MaxObjectNameLength)
	}
	n, err := rpc.NewObjectName(prefix)
	if err != nil {
		return "", fmt.Errorf("qgyolaus: profile: %w", err)
	}
	if generic {
		return n.String() + "*", nil
	}
	return n.String(), nil
}

func fixed(codec ccsid.Codec, s string, width int) ([]byte, error) {
	b, err := ccsid.EncodeFixed(codec, s, width)
	if err != nil {
		return nil, fmt.Errorf("qgyolaus: %w", err)
	}
	return b, nil
}

func input(b []byte) rpc.Parameter {
	p, _ := rpc.Input(b) // small fixed-size inputs
	return p
}

// NewCall builds the QGYOLAUS call for format. codec encodes the
// character parameters (the system CCSID).
//
// Parameters, in order: receiver (zero-length input, the JTOpen 4.8 wire
// form), receiver length 0, list information (output 80), records to return
// -1, format name, selection, group, error code, profile.
func NewCall[T any](format Format[T], req Request, codec ccsid.Codec) (rpc.Call, error) {
	if format.IsZero() {
		return rpc.Call{}, errZeroFormat
	}
	sel, err := req.Selection.value()
	if err != nil {
		return rpc.Call{}, err
	}
	group := "*NONE"
	switch {
	case req.Selection == SelectMembers:
		g, err := rpc.NewObjectName(req.Group)
		if err != nil {
			return rpc.Call{}, fmt.Errorf("qgyolaus: group: %w", err)
		}
		group = g.String()
	case req.Group != "":
		return rpc.Call{}, errors.New("qgyolaus: Group is only valid with SelectMembers")
	}
	profile, err := validateProfile(req.Profile)
	if err != nil {
		return rpc.Call{}, err
	}
	name, err := fixed(codec, format.name, 8)
	if err != nil {
		return rpc.Call{}, err
	}
	selBytes, err := fixed(codec, sel, 10)
	if err != nil {
		return rpc.Call{}, err
	}
	groupBytes, err := fixed(codec, group, 10)
	if err != nil {
		return rpc.Call{}, err
	}
	profileBytes, err := fixed(codec, profile, 10)
	if err != nil {
		return rpc.Call{}, err
	}
	listInfo, _ := rpc.Output(openlist.Length)
	params := []rpc.Parameter{
		rpc.EmptyInput(),
		input(qgy.Binary4(0)),
		listInfo,
		input(qgy.Binary4(synchronous)),
		input(name),
		input(selBytes),
		input(groupBytes),
		req.ErrorCode.RPC(),
		input(profileBytes),
	}
	option := req.MessageOption
	if option == 0 {
		option = rpc.MessagesUpToTen
	}
	return rpc.NewCall(qgy.LibraryOr(req.Library), Program(), params, option)
}

// OpenResult is a validated open.
type OpenResult struct {
	Info openlist.Info
	// Handle identifies the list on this server job only.
	Handle openlist.Handle
}

// ErrMetadata is matched when an open succeeded on the server but its list
// information fails validation. The handle may still have been recovered;
// the caller must then close it.
var ErrMetadata = errors.New("qgyolaus: invalid list information")

// ParseOpen validates the result of a NewCall call. On a metadata failure it
// returns an error matching ErrMetadata together with an OpenResult whose
// Handle is set when it could be recovered (handleRecovered), so the caller
// can still close the list. jobCodec decodes an error-code exception ID.
func ParseOpen[T any](format Format[T], call rpc.Call, req Request, res *rpc.Result, jobCodec ccsid.Codec) (result *OpenResult, handleRecovered bool, err error) {
	if format.IsZero() {
		return nil, false, errZeroFormat
	}
	if err := res.Validate(call); err != nil {
		return nil, false, err
	}
	if call.ParameterCount() != parameterCount {
		return nil, false, fmt.Errorf("qgyolaus: call has %d parameters, want %d", call.ParameterCount(), parameterCount)
	}
	if err := req.ErrorCode.Check(call.Library(), call.Program(), res.Parameters[paramErrorCode], jobCodec); err != nil {
		return nil, false, err
	}
	raw, ok := res.Output(paramListInfo)
	if !ok {
		return nil, false, fmt.Errorf("%w: list information was not returned", ErrMetadata)
	}
	handle, recovered := openlist.RecoverHandle(raw)
	result = &OpenResult{Handle: handle}
	info, err := openlist.Unmarshal(raw)
	if err != nil {
		return result, recovered, fmt.Errorf("%w: %w", ErrMetadata, err)
	}
	result.Info = info
	switch {
	case info.Status != openlist.StatusComplete:
		// The synchronous (-1) request is complete when it returns.
		err = fmt.Errorf("%w: list status %s, want complete", ErrMetadata, info.Status)
	case info.RecordsReturned != 0:
		err = fmt.Errorf("%w: %d records returned into a zero-length receiver", ErrMetadata, info.RecordsReturned)
	case info.RecordLength != 0 && uint32(info.RecordLength) != format.length:
		// A record length of 0 is accepted at open: no record was returned.
		err = fmt.Errorf("%w: record length %d, %s is %d bytes", ErrMetadata, info.RecordLength, format.name, format.length)
	}
	return result, recovered, err
}

// Open runs QGYOLAUS through caller. See ParseOpen for the error contract:
// when handleRecovered is true the list exists on the server and must be
// closed even though err is non-nil.
func Open[T any](ctx context.Context, caller rpc.Caller, format Format[T], req Request, systemCodec, jobCodec ccsid.Codec) (result *OpenResult, handleRecovered bool, err error) {
	call, err := NewCall(format, req, systemCodec)
	if err != nil {
		return nil, false, err
	}
	res, err := caller.Call(ctx, call)
	if err != nil {
		return nil, false, err
	}
	return ParseOpen(format, call, req, res, jobCodec)
}
