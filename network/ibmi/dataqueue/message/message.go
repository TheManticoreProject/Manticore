// Package message encodes and decodes the Data Queue host-server requests
// and replies (server ID 0xE007): exchange attributes (0x0000/0x8000),
// query attributes (0x0001/0x8001), read and peek (0x0002/0x8003), create
// (0x0003), delete (0x0004), write (0x0005) and clear (0x0006), the last
// four answered by the common reply 0x8002.
//
// Layouts follow the JTOpen 4.8 generation that the V5R4 baseline targets,
// as recorded in iSeriesNavigatorReverse docs/16-data-queue-host-server.md,
// whose constructor vectors the golden tests reproduce byte for byte. Queue
// and library names arrive here already encoded as 10-byte, blank-padded
// EBCDIC; entry and key bytes are opaque and never transcoded.
package message

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/startup"
)

// Request IDs.
const (
	RequestExchangeAttributes uint16 = 0x0000
	RequestQueryAttributes    uint16 = 0x0001
	RequestRead               uint16 = 0x0002
	RequestCreate             uint16 = 0x0003
	RequestDelete             uint16 = 0x0004
	RequestWrite              uint16 = 0x0005
	RequestClear              uint16 = 0x0006
)

// Reply IDs.
const (
	ReplyExchangeAttributes uint16 = 0x8000
	ReplyQueryAttributes    uint16 = 0x8001
	ReplyCommon             uint16 = 0x8002
	ReplyRead               uint16 = 0x8003
)

// Code points of the optional LL/CP fields.
const (
	CPEntry uint16 = 0x5001
	CPKey   uint16 = 0x5002
)

// Limits enforced by the V5R4-era client.
const (
	MaxEntryLength  = 64512
	MaxKeyLength    = 256
	DescriptionSize = 50
	NameSize        = 10
)

// ClientVersion advertises 64-KiB queue support in exchange attributes.
const ClientVersion uint32 = 1

// EBCDIC '0' and '1', the encoding of every boolean in this protocol.
const (
	False byte = 0xF0
	True  byte = 0xF1
)

func flag(b bool) byte {
	if b {
		return True
	}
	return False
}

// Object names a queue: the 10-byte blank-padded EBCDIC queue and library.
type Object struct {
	Queue, Library [NameSize]byte
}

func (o Object) append(b []byte) []byte { return append(append(b, o.Queue[:]...), o.Library[:]...) }

func frame(id uint16, corr uint32, template, variable []byte) hostserver.Frame {
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDDataQueue, CorrelationID: corr, RequestID: id},
		Template: template, Variable: variable,
	}
}

// ExchangeAttributesFrame encodes request 0x0000: client version 1 and
// data-stream level 0.
func ExchangeAttributesFrame(corr uint32) hostserver.Frame {
	return frame(RequestExchangeAttributes, corr, append(binary.BigEndian.AppendUint32(nil, ClientVersion), 0, 0), nil)
}

// QueryAttributesFrame encodes request 0x0001.
func QueryAttributesFrame(corr uint32, o Object) hostserver.Frame {
	return frame(RequestQueryAttributes, corr, o.append(nil), nil)
}

// DeleteFrame encodes request 0x0004.
func DeleteFrame(corr uint32, o Object) hostserver.Frame {
	return frame(RequestDelete, corr, o.append(nil), nil)
}

// ReadRequest is request 0x0002.
type ReadRequest struct {
	Correlation uint32
	Object      Object
	// Keyed selects a keyed search; Search is then the encoded two-byte
	// operator (EQ, NE, LT, LE, GT, GE) and Key the comparison key.
	Keyed  bool
	Search [2]byte
	Key    []byte
	// Wait is in seconds; -1 waits indefinitely.
	Wait int32
	// Peek leaves the entry on the queue.
	Peek bool
}

// Frame encodes the request: a 28-byte template, then CP 0x5002 key for a
// keyed search.
func (r ReadRequest) Frame() (hostserver.Frame, error) {
	if r.Wait < -1 {
		return hostserver.Frame{}, fmt.Errorf("message: wait %d below -1", r.Wait)
	}
	t := r.Object.append(make([]byte, 0, 28))
	t = append(t, flag(r.Keyed))
	var v []byte
	if r.Keyed {
		if err := checkKey(r.Key); err != nil {
			return hostserver.Frame{}, err
		}
		t = append(t, r.Search[:]...)
		v, _ = hostserver.AppendField(nil, CPKey, r.Key)
	} else {
		if len(r.Key) != 0 {
			return hostserver.Frame{}, fmt.Errorf("message: key on a non-keyed read")
		}
		t = append(t, 0, 0)
	}
	t = binary.BigEndian.AppendUint32(t, uint32(r.Wait))
	t = append(t, flag(r.Peek))
	return frame(RequestRead, r.Correlation, t, v), nil
}

func checkKey(key []byte) error {
	if len(key) == 0 || len(key) > MaxKeyLength {
		return fmt.Errorf("message: key of %d bytes outside 1-%d", len(key), MaxKeyLength)
	}
	return nil
}

// WriteRequest is request 0x0005.
type WriteRequest struct {
	Correlation uint32
	Object      Object
	Entry       []byte
	// Key, when non-nil, makes a keyed write.
	Key []byte
}

// Frame encodes the request: a 22-byte template (keyed indicator, want
// reply 1), CP 0x5001 entry, then CP 0x5002 key for a keyed write.
func (r WriteRequest) Frame() (hostserver.Frame, error) {
	if len(r.Entry) > MaxEntryLength {
		return hostserver.Frame{}, fmt.Errorf("message: entry of %d bytes exceeds %d", len(r.Entry), MaxEntryLength)
	}
	keyed := r.Key != nil
	t := append(r.Object.append(make([]byte, 0, 22)), flag(keyed), True)
	v, _ := hostserver.AppendField(nil, CPEntry, r.Entry)
	if keyed {
		if err := checkKey(r.Key); err != nil {
			return hostserver.Frame{}, err
		}
		v, _ = hostserver.AppendField(v, CPKey, r.Key)
	}
	return frame(RequestWrite, r.Correlation, t, v), nil
}

// ClearFrame encodes request 0x0006: a 21-byte template, then CP 0x5002
// key when key is non-nil (clear only the entries matching it).
func ClearFrame(corr uint32, o Object, key []byte) (hostserver.Frame, error) {
	keyed := key != nil
	t := append(o.append(make([]byte, 0, 21)), flag(keyed))
	var v []byte
	if keyed {
		if err := checkKey(key); err != nil {
			return hostserver.Frame{}, err
		}
		v, _ = hostserver.AppendField(nil, CPKey, key)
	}
	return frame(RequestClear, corr, t, v), nil
}

// Queue types on the wire: create uses the EBCDIC digit, the query reply
// the low nibble.
const (
	TypeFIFO  byte = 0
	TypeLIFO  byte = 1
	TypeKeyed byte = 2
)

// Public authority codes for create.
const (
	AuthorityAll       byte = 0xF0
	AuthorityChange    byte = 0xF1
	AuthorityExclude   byte = 0xF2
	AuthorityUse       byte = 0xF3
	AuthorityLibCrtAut byte = 0xF4
)

// CreateRequest is request 0x0003.
type CreateRequest struct {
	Correlation uint32
	Object      Object
	MaxEntry    uint32
	Authority   byte
	SaveSender  bool
	Type        byte
	// KeyLength is 1-256 for a keyed queue and 0 otherwise.
	KeyLength       uint16
	ForceAuxStorage bool
	// Description is the encoded, blank-padded text.
	Description [DescriptionSize]byte
}

// Frame encodes the fixed 100-byte request (template length 80).
func (r CreateRequest) Frame() (hostserver.Frame, error) {
	switch {
	case r.MaxEntry < 1 || r.MaxEntry > MaxEntryLength:
		return hostserver.Frame{}, fmt.Errorf("message: maximum entry length %d outside 1-%d", r.MaxEntry, MaxEntryLength)
	case r.Authority < AuthorityAll || r.Authority > AuthorityLibCrtAut:
		return hostserver.Frame{}, fmt.Errorf("message: undefined authority code 0x%02X", r.Authority)
	case r.Type > TypeKeyed:
		return hostserver.Frame{}, fmt.Errorf("message: undefined queue type %d", r.Type)
	case (r.Type == TypeKeyed) != (r.KeyLength != 0) || r.KeyLength > MaxKeyLength:
		return hostserver.Frame{}, fmt.Errorf("message: key length %d invalid for queue type %d", r.KeyLength, r.Type)
	}
	t := r.Object.append(make([]byte, 0, 80))
	t = binary.BigEndian.AppendUint32(t, r.MaxEntry)
	t = append(t, r.Authority, flag(r.SaveSender), 0xF0|r.Type)
	t = binary.BigEndian.AppendUint16(t, r.KeyLength)
	t = append(t, flag(r.ForceAuxStorage))
	t = append(t, r.Description[:]...)
	return frame(RequestCreate, r.Correlation, t, nil), nil
}

// body returns a reply's bytes after the header, whatever its declared
// template length.
func body(f hostserver.Frame) []byte { return startup.Body(f) }

func short(what string, have, want int) error { return startup.Short(what, have, want) }

// Return codes of the common reply 0x8002.
const (
	RCSuccess              uint16 = 0xF000
	RCCommandCheck         uint16 = 0xF001
	RCProtocolError        uint16 = 0xF002
	RCSyntaxError          uint16 = 0xF003
	RCQueueDestroyed       uint16 = 0xF004
	RCUnsupportedLength    uint16 = 0xF005
	RCNoData               uint16 = 0xF006
	RCInvalidLevel         uint16 = 0xF007
	RCInvalidVRM           uint16 = 0xF008
	RCExitRejected         uint16 = 0xF009
	RCExitNotAuthorized    uint16 = 0xF00A
	RCExitNotFound         uint16 = 0xF00B
	RCExitFailed           uint16 = 0xF00D
	RCExitInvalidProgramNo uint16 = 0xF00E
)

// CommonReply is reply 0x8002.
type CommonReply struct {
	ReturnCode uint16
	// Message is the raw value of the optional LL/CP message field (an IBM
	// message, encoded in the job CCSID); MessageCodePoint is its code
	// point, which the protocol does not define. HasMessage is false when
	// the reply has no field.
	Message          []byte
	MessageCodePoint uint16
	HasMessage       bool
}

// ParseCommonReply decodes reply 0x8002: a 2-byte return code, then at most
// one LL/CP message field that must end the frame.
func ParseCommonReply(f hostserver.Frame, maxMessage int) (*CommonReply, error) {
	b := body(f)
	if len(b) < 2 {
		return nil, short("0x8002 reply", len(b), 2)
	}
	r := &CommonReply{ReturnCode: binary.BigEndian.Uint16(b)}
	if len(b) == 2 {
		return r, nil
	}
	fields, err := hostserver.ParseFields(b[2:], hostserver.FieldPolicy{MaxFields: 1})
	if err != nil {
		return nil, err
	}
	if len(fields[0].Value) > maxMessage {
		return nil, &hostserver.MalformedError{What: "field", Offset: hostserver.HeaderLength + 2, Reason: fmt.Sprintf("message of %d bytes exceeds the %d-byte limit", len(fields[0].Value), maxMessage)}
	}
	r.Message, r.MessageCodePoint, r.HasMessage = append([]byte(nil), fields[0].Value...), fields[0].CodePoint, true
	return r, nil
}

// AttributesReply is reply 0x8001.
type AttributesReply struct {
	MaxEntry        uint32
	SaveSender      bool
	Type            byte
	KeyLength       uint16
	ForceAuxStorage bool
	// Description is the raw 50-byte text in the system CCSID.
	Description [DescriptionSize]byte
}

// ParseAttributesReply decodes reply 0x8001 (at least 61 body bytes).
func ParseAttributesReply(f hostserver.Frame) (*AttributesReply, error) {
	b := body(f)
	if len(b) < 61 {
		return nil, short("0x8001 reply", len(b), 61)
	}
	r := &AttributesReply{
		MaxEntry:        binary.BigEndian.Uint32(b[2:]),
		SaveSender:      b[6] == True,
		Type:            b[7] & 0x0F,
		KeyLength:       binary.BigEndian.Uint16(b[8:]),
		ForceAuxStorage: b[10] == True,
	}
	if r.Type > TypeKeyed {
		return nil, &hostserver.MalformedError{What: "0x8001 reply", Offset: hostserver.HeaderLength + 7, Reason: fmt.Sprintf("undefined queue type %d", r.Type)}
	}
	copy(r.Description[:], b[11:61])
	return r, nil
}

// SenderInfoSize is the size of the sender information in a read reply.
const SenderInfoSize = 36

// ReadReply is reply 0x8003.
type ReadReply struct {
	// Sender is the raw sender information; HasSender is false when it is
	// all EBCDIC blanks.
	Sender    [SenderInfoSize]byte
	HasSender bool
	Entry     []byte
	Key       []byte
	HasKey    bool
}

// ParseReadReply decodes reply 0x8003: a 2-byte status, 36 bytes of sender
// information, then the CP 0x5001 entry and an optional CP 0x5002 key.
// Duplicate fields are rejected; returned slices are copies.
func ParseReadReply(f hostserver.Frame) (*ReadReply, error) {
	b := body(f)
	if len(b) < 2+SenderInfoSize {
		return nil, short("0x8003 reply", len(b), 2+SenderInfoSize)
	}
	fields, err := hostserver.ParseFields(b[2+SenderInfoSize:], hostserver.FieldPolicy{MaxFields: 4})
	if err != nil {
		return nil, err
	}
	r := &ReadReply{}
	copy(r.Sender[:], b[2:])
	for _, c := range r.Sender {
		if c != 0x40 {
			r.HasSender = true
			break
		}
	}
	entry, ok := fields.Get(CPEntry)
	if !ok {
		return nil, &hostserver.MalformedError{What: "0x8003 reply", Offset: hostserver.HeaderLength + 2 + SenderInfoSize, Reason: "no entry field (CP 0x5001)"}
	}
	r.Entry = append([]byte{}, entry...)
	if key, ok := fields.Get(CPKey); ok {
		r.Key, r.HasKey = append([]byte{}, key...), true
	}
	return r, nil
}
