// Package message encodes and decodes the File host-server requests and
// replies (server ID 0xE002) of the classic, data-stream level 8 read path:
// exchange attributes (0x0016/0x8009), open (0x0002/0x8002), read
// (0x0003/0x8003, possibly chained), close (0x0009/0x8004) and the generic
// return-code reply 0x8001.
//
// Every file-server packet starts its template with a two-byte chain
// indicator; bit 0 set means another reply with the same correlation
// follows. Layouts follow the JTOpen 4.8 generation that the V5R4 baseline
// targets, as recorded in iSeriesNavigatorReverse
// docs/08-file-server-ifs.md and docs/10, whose constructor vectors the
// golden tests reproduce byte for byte. The level-16 (64-bit offset)
// layouts of later releases are not implemented.
//
// Observed on V5R4 (2026-10-15): the server answers a level-8 request with
// level 12 and pathname CCSID 1200; the read reply's data field uses code
// point 0x0005 (kept, never checked); end of file is an empty 0x8003; a
// read of 4 MiB came back as a single packet; close answers 0x8004 with
// the return code followed by four zero bytes.
package message

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/startup"
)

// Request IDs.
const (
	RequestExchangeAttributes uint16 = 0x0016
	RequestOpen               uint16 = 0x0002
	RequestRead               uint16 = 0x0003
	RequestClose              uint16 = 0x0009
)

// Reply IDs. They are not request|0x8000.
const (
	ReplyReturnCode         uint16 = 0x8001
	ReplyOpen               uint16 = 0x8002
	ReplyRead               uint16 = 0x8003
	ReplyClose              uint16 = 0x8004
	ReplyExchangeAttributes uint16 = 0x8009
)

// Code points.
const (
	CPCCSIDList uint16 = 0x000A
	CPFileName  uint16 = 0x0002
)

// DataStreamLevel is the level the V5R3-V5R4 client requests and the only
// one whose layouts are implemented.
const DataStreamLevel uint16 = 8

// Exchange-attributes flags: GMT timestamps and PC pattern matching. The
// POSIX return-code bit (0x0008) stays clear.
const (
	FlagGMT       uint16 = 0x0004
	FlagPCPattern uint16 = 0x0002
)

// Pathname CCSIDs the client prefers, in order.
var PreferredCCSIDs = []uint16{1200, 13488, 61952}

// Chain is bit 0 of the chain indicator.
const Chain uint16 = 0x0001

// MaxOffset bounds read offsets below data-stream level 16.
const MaxOffset = 0x7FFFFFFF

func frame(id uint16, corr uint32, template, variable []byte) hostserver.Frame {
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDFile, CorrelationID: corr, RequestID: id},
		Template: template, Variable: variable,
	}
}

// ExchangeAttributesFrame encodes request 0x0016: level, flags, maximum
// data block 0xFFFFFFFF, then the preferred pathname CCSIDs.
func ExchangeAttributesFrame(corr uint32, level uint16, ccsids []uint16) (hostserver.Frame, error) {
	if len(ccsids) == 0 || len(ccsids) > 16 {
		return hostserver.Frame{}, fmt.Errorf("message: %d preferred CCSIDs, need 1-16", len(ccsids))
	}
	t := binary.BigEndian.AppendUint16(make([]byte, 2, 10), level)
	t = binary.BigEndian.AppendUint16(t, FlagGMT|FlagPCPattern)
	t = binary.BigEndian.AppendUint32(t, 0xFFFFFFFF)
	var list []byte
	for _, c := range ccsids {
		list = binary.BigEndian.AppendUint16(list, c)
	}
	v, _ := hostserver.AppendField(nil, CPCCSIDList, list)
	return frame(RequestExchangeAttributes, corr, t, v), nil
}

// Open access intents, deny modes, conversions and duplicate options.
const (
	AccessRead        uint16 = 1
	DenyNone          uint16 = 0
	ConvertNone       uint16 = 0
	OpenExistingOnly  uint16 = 8 // fail if absent, open if present
	WorkingDirectory  uint32 = 1 // path-based open
	AttributeListBase uint16 = 1
)

// OpenRequest is request 0x0002 in its level-below-16 form.
type OpenRequest struct {
	Correlation uint32
	PathCCSID   uint16
	// DataCCSID is the requested file-data CCSID; with ConvertNone the
	// bytes are returned unchanged.
	DataCCSID  uint16
	Access     uint16
	Deny       uint16
	Conversion uint16
	Duplicate  uint16
	// Path is the encoded pathname, with no terminator.
	Path []byte
}

// Frame encodes the request: a 36-byte template, then CP 0x0002 path.
func (r OpenRequest) Frame() (hostserver.Frame, error) {
	if len(r.Path) == 0 {
		return hostserver.Frame{}, fmt.Errorf("message: empty path")
	}
	t := make([]byte, 2, 36)
	t = binary.BigEndian.AppendUint16(t, r.PathCCSID)
	t = binary.BigEndian.AppendUint32(t, WorkingDirectory)
	t = binary.BigEndian.AppendUint16(t, r.DataCCSID)
	t = binary.BigEndian.AppendUint16(t, r.Access)
	t = binary.BigEndian.AppendUint16(t, r.Deny)
	t = binary.BigEndian.AppendUint16(t, r.Conversion)
	t = binary.BigEndian.AppendUint16(t, r.Duplicate)
	t = append(t, 0, 0, 0, 0, 0, 0, 0, 0) // create size, fixed attributes
	t = binary.BigEndian.AppendUint16(t, AttributeListBase)
	t = append(t, 0, 0, 0, 0, 0, 0, 0, 0) // pre-read offset and length
	v, err := hostserver.AppendField(nil, CPFileName, r.Path)
	if err != nil {
		return hostserver.Frame{}, err
	}
	return frame(RequestOpen, r.Correlation, t, v), nil
}

// ReadFrame encodes request 0x0003 in its level-below-16 form: handle, base
// offset 0, relative offset, length and pre-read length 0.
func ReadFrame(corr, handle uint32, offset int64, length uint32) (hostserver.Frame, error) {
	if offset < 0 || offset > MaxOffset {
		return hostserver.Frame{}, fmt.Errorf("message: offset %d outside 0-%d", offset, MaxOffset)
	}
	if length == 0 {
		return hostserver.Frame{}, fmt.Errorf("message: zero read length")
	}
	t := binary.BigEndian.AppendUint32(make([]byte, 2, 22), handle)
	t = binary.BigEndian.AppendUint32(t, 0)
	t = binary.BigEndian.AppendUint32(t, uint32(offset))
	t = binary.BigEndian.AppendUint32(t, length)
	t = binary.BigEndian.AppendUint32(t, 0)
	return frame(RequestRead, corr, t, nil), nil
}

// CloseFrame encodes request 0x0009 for a read-only close: data flags 2,
// CCSID 0xFFFF, amount accessed 100, access history 0, modify date 0.
func CloseFrame(corr, handle uint32) hostserver.Frame {
	t := binary.BigEndian.AppendUint32(make([]byte, 2, 21), handle)
	t = append(t, 0x00, 0x02, 0xFF, 0xFF, 0x00, 0x64, 0x00)
	t = append(t, 0, 0, 0, 0, 0, 0, 0, 0)
	return frame(RequestClose, corr, t, nil)
}

// Body returns a reply's bytes after the header, whatever its declared
// template length.
func Body(f hostserver.Frame) []byte { return startup.Body(f) }

func short(what string, have, want int) error { return startup.Short(what, have, want) }

// Chained reports whether another reply with the same correlation follows
// f. It is false for a body too short to hold the indicator.
func Chained(f hostserver.Frame) bool {
	b := Body(f)
	return len(b) >= 2 && binary.BigEndian.Uint16(b)&Chain != 0
}

// ParseReturnCode decodes the generic reply 0x8001 or the close reply
// 0x8004: chain indicator, then a two-byte return code.
func ParseReturnCode(f hostserver.Frame) (uint16, error) {
	b := Body(f)
	if len(b) < 4 {
		return 0, short(fmt.Sprintf("0x%04X reply", f.Header.RequestID), len(b), 4)
	}
	return binary.BigEndian.Uint16(b[2:]), nil
}

// AttributesReply is reply 0x8009.
type AttributesReply struct {
	Level    uint16
	MaxBlock uint32
	// CCSIDs lists the server's pathname CCSIDs; the first is the one
	// selected.
	CCSIDs []uint16
}

// ParseAttributesReply decodes reply 0x8009: chain, level, an
// uninterpreted field, the maximum data block, then the CCSID-list LL/CP
// field, which must hold at least one CCSID and fit the frame.
func ParseAttributesReply(f hostserver.Frame) (*AttributesReply, error) {
	b := Body(f)
	if len(b) < 18 {
		return nil, short("0x8009 reply", len(b), 18)
	}
	ll := binary.BigEndian.Uint32(b[10:])
	if ll < 8 || uint64(ll) > uint64(len(b)-10) {
		return nil, &hostserver.MalformedError{What: "0x8009 reply", Offset: hostserver.HeaderLength + 10, Reason: fmt.Sprintf("CCSID list length %d outside 8-%d", ll, len(b)-10)}
	}
	r := &AttributesReply{Level: binary.BigEndian.Uint16(b[2:]), MaxBlock: binary.BigEndian.Uint32(b[6:])}
	for i := 16; i+2 <= 10+int(ll); i += 2 {
		r.CCSIDs = append(r.CCSIDs, binary.BigEndian.Uint16(b[i:]))
	}
	return r, nil
}

// OpenReplyLength is the level-8 0x8002 body length (through the access
// history byte at frame offset 88).
const OpenReplyLength = 69

// OpenReply is reply 0x8002 at level 8.
type OpenReply struct {
	Handle    uint32
	FileID    uint64
	DataCCSID uint16
	Action    uint16
	Created   time.Time
	Modified  time.Time
	Accessed  time.Time
	// Size is the 32-bit level-8 file size, unsigned.
	Size            uint32
	FixedAttributes uint32
	Version         uint32
}

func timestamp(b []byte) time.Time {
	return time.Unix(int64(binary.BigEndian.Uint32(b)), int64(binary.BigEndian.Uint32(b[4:]))*1000).UTC()
}

// ParseOpenReply decodes a level-8 reply 0x8002. Success has no return
// code: the reply ID is the signal.
func ParseOpenReply(f hostserver.Frame) (*OpenReply, error) {
	b := Body(f)
	if len(b) < OpenReplyLength {
		return nil, short("0x8002 reply", len(b), OpenReplyLength)
	}
	return &OpenReply{
		Handle:          binary.BigEndian.Uint32(b[2:]),
		FileID:          binary.BigEndian.Uint64(b[6:]),
		DataCCSID:       binary.BigEndian.Uint16(b[14:]),
		Action:          binary.BigEndian.Uint16(b[16:]),
		Created:         timestamp(b[18:]),
		Modified:        timestamp(b[26:]),
		Accessed:        timestamp(b[34:]),
		Size:            binary.BigEndian.Uint32(b[42:]),
		FixedAttributes: binary.BigEndian.Uint32(b[46:]),
		Version:         binary.BigEndian.Uint32(b[62:]),
	}, nil
}

// ReadReply is one reply 0x8003 of a possibly chained read.
type ReadReply struct {
	Chained   bool
	DataCCSID uint16
	// CodePoint is the data field's code point, which the protocol leaves
	// unvalidated; it is kept for diagnostics.
	CodePoint uint16
	Data      []byte
}

// ParseReadReply decodes reply 0x8003: chain, data CCSID, then one data
// LL/CP field that must end the frame. Data is a copy.
func ParseReadReply(f hostserver.Frame) (*ReadReply, error) {
	b := Body(f)
	if len(b) < 10 {
		return nil, short("0x8003 reply", len(b), 10)
	}
	ll := binary.BigEndian.Uint32(b[4:])
	if ll < 6 || uint64(ll) != uint64(len(b)-4) {
		return nil, &hostserver.MalformedError{What: "0x8003 reply", Offset: hostserver.HeaderLength + 4, Reason: fmt.Sprintf("data length %d does not match the %d remaining bytes", ll, len(b)-4)}
	}
	return &ReadReply{
		Chained:   binary.BigEndian.Uint16(b)&Chain != 0,
		DataCCSID: binary.BigEndian.Uint16(b[2:]),
		CodePoint: binary.BigEndian.Uint16(b[8:]),
		Data:      append([]byte{}, b[10:]...),
	}, nil
}
