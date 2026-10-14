package message

import (
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// ExchangeAttributesRequest is request 0x7003. A password sign-on always
// sends a client seed; nil omits CP 0x1103 for token paths.
type ExchangeAttributesRequest struct {
	ClientSeed *[8]byte
}

// Frame encodes the request: client version 1 and data-stream level 2.
func (r ExchangeAttributesRequest) Frame() hostserver.Frame {
	fields := []hostserver.Field{
		{CodePoint: CPVersion, Value: u32(ClientVersion)},
		{CodePoint: CPDataStreamLevel, Value: u16(ClientDataStreamLevel)},
	}
	if r.ClientSeed != nil {
		fields = append(fields, hostserver.Field{CodePoint: CPSeed, Value: r.ClientSeed[:]})
	}
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDSignon, RequestID: RequestExchangeAttributes},
		Variable: appendFields(fields...),
	}
}

// ExchangeAttributesReply is reply 0xF003.
type ExchangeAttributesReply struct {
	ReturnCode uint32
	// ServerVersion is the raw server VRM integer (CP 0x1101).
	ServerVersion uint32
	// ServerLevel is the Sign-on server data-stream level (CP 0x1102).
	ServerLevel uint16
	ServerSeed  [8]byte
	// PasswordLevel is the raw CP 0x1119 byte; HasPasswordLevel is false when
	// the server omitted it.
	PasswordLevel    byte
	HasPasswordLevel bool
	// Fields holds every reply field, including ones not interpreted here.
	Fields hostserver.Fields
}

// ParseExchangeAttributesReply decodes reply 0xF003. With a zero return code
// the server version, level and seed are required. A nonzero return code is
// returned with whatever fields were present; the caller classifies it.
func ParseExchangeAttributesReply(f hostserver.Frame) (*ExchangeAttributesReply, error) {
	rc, fields, err := replyBody(f, CPVersion, CPDataStreamLevel, CPSeed, CPPasswordLevel)
	if err != nil {
		return nil, err
	}
	r := &ExchangeAttributesReply{ReturnCode: rc, Fields: fields}
	if rc != 0 {
		return r, nil
	}
	v, ok, err := fixedField(fields, CPVersion, 4)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, missing(CPVersion)
	}
	r.ServerVersion = binary.BigEndian.Uint32(v)
	if v, ok, err = fixedField(fields, CPDataStreamLevel, 2); err != nil {
		return nil, err
	} else if !ok {
		return nil, missing(CPDataStreamLevel)
	}
	r.ServerLevel = binary.BigEndian.Uint16(v)
	if v, ok, err = fixedField(fields, CPSeed, 8); err != nil {
		return nil, err
	} else if !ok {
		return nil, missing(CPSeed)
	}
	copy(r.ServerSeed[:], v)
	if v, ok, err = fixedField(fields, CPPasswordLevel, 1); err != nil {
		return nil, err
	} else if ok {
		r.PasswordLevel, r.HasPasswordLevel = v[0], true
	}
	return r, nil
}
