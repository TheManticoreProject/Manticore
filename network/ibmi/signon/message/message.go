// Package message encodes and decodes the Sign-on host-server requests and
// replies used for validation and tokens: exchange attributes
// (0x7003/0xF003), retrieve sign-on information (0x7004/0xF004), generate
// profile token (0x7007/0xF007, iSeriesNavigatorReverse docs/21) and end
// server (0x7006).
//
// Layouts follow the JTOpen 4.8 generation the V5R4 baseline targets
// (client data-stream level 2, client CCSID 13488), as recorded in
// iSeriesNavigatorReverse docs/05-signon-and-authentication.md and the
// golden request frames in docs/10-v5r4-compatibility-and-golden-vectors.md.
//
// Replies carry a 4-byte return code immediately after the header, then
// LL/CP fields. The decoders read the return code from the first four body
// bytes whatever the declared template length, as the validated proof of
// concept did.
package message

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// Request and reply IDs.
const (
	RequestExchangeAttributes uint16 = 0x7003
	ReplyExchangeAttributes   uint16 = 0xF003
	RequestRetrieveInfo       uint16 = 0x7004
	ReplyRetrieveInfo         uint16 = 0xF004
	RequestEndServer          uint16 = 0x7006
	RequestGenerateToken      uint16 = 0x7007
	ReplyGenerateToken        uint16 = 0xF007
)

// Code points.
const (
	CPVersion            uint16 = 0x1101
	CPDataStreamLevel    uint16 = 0x1102
	CPSeed               uint16 = 0x1103
	CPUserID             uint16 = 0x1104
	CPPasswordSubstitute uint16 = 0x1105
	CPClientCCSID        uint16 = 0x1113
	CPServerCCSID        uint16 = 0x1114
	CPPasswordLevel      uint16 = 0x1119
	CPJobIdentity        uint16 = 0x111F
	CPAuthToken          uint16 = 0x1115
	CPTokenType          uint16 = 0x1116
	CPTokenTimeout       uint16 = 0x1117
)

// Values sent by the JTOpen 4.8 generation.
const (
	ClientVersion         uint32 = 1
	ClientDataStreamLevel uint16 = 2
	// ClientCCSID is UTF-16 (13488); current JTOpen sends 1200.
	ClientCCSID uint32 = 13488
)

// maxReplyFields bounds the LL/CP fields accepted in one reply.
const maxReplyFields = 64

func u16(v uint16) []byte { return []byte{byte(v >> 8), byte(v)} }
func u32(v uint32) []byte { return binary.BigEndian.AppendUint32(nil, v) }

func appendFields(fields ...hostserver.Field) []byte {
	var out []byte
	for _, f := range fields {
		out, _ = hostserver.AppendField(out, f.CodePoint, f.Value) // values are small and fixed
	}
	return out
}

// replyBody splits a reply into its return code and LL/CP fields. consumed
// lists the code points the caller reads: they must not repeat, while other
// code points may.
func replyBody(f hostserver.Frame, consumed ...uint16) (uint32, hostserver.Fields, error) {
	body := make([]byte, 0, len(f.Template)+len(f.Variable))
	body = append(append(body, f.Template...), f.Variable...)
	if len(body) < 4 {
		return 0, nil, &hostserver.MalformedError{What: "reply", Offset: hostserver.HeaderLength, Reason: fmt.Sprintf("%d-byte body cannot hold a return code", len(body))}
	}
	singletons := map[uint16]bool{}
	for _, cp := range consumed {
		singletons[cp] = true
	}
	fields, err := hostserver.ParseFields(body[4:], hostserver.FieldPolicy{MaxFields: maxReplyFields, Singletons: singletons})
	if err != nil {
		return 0, nil, err
	}
	return binary.BigEndian.Uint32(body), fields, nil
}

// fixedField returns the value of cp, which must be exactly size bytes.
func fixedField(fields hostserver.Fields, cp uint16, size int) ([]byte, bool, error) {
	v, ok := fields.Get(cp)
	if !ok {
		return nil, false, nil
	}
	if len(v) != size {
		return nil, true, &hostserver.MalformedError{What: "field", Reason: fmt.Sprintf("code point 0x%04X has %d bytes, want %d", cp, len(v), size)}
	}
	return v, true, nil
}

func missing(cp uint16) error {
	return &hostserver.MalformedError{What: "reply", Reason: fmt.Sprintf("required code point 0x%04X is missing", cp)}
}
