package message

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// MessageRecordMinLength is the fixed part of a classic message record.
const MessageRecordMinLength = 41

// MessageRecord is one classic (V5R4 / JTOpen 4.8) IBM message record with
// its character fields still in the job CCSID.
type MessageRecord struct {
	// Tag is the 2-byte format tag; JTOpen 4.8 ignores it and its V5R4 value
	// is not yet established, so it is only recorded.
	Tag      uint16
	ID       [7]byte
	Type     uint8
	Severity uint16
	File     [10]byte
	Library  [10]byte
	// Substitution and Text are copies, safe to retain.
	Substitution []byte
	Text         []byte
	// Extension holds record bytes after the text, if any.
	Extension []byte
}

// MessageLimits bounds message parsing before allocation.
type MessageLimits struct {
	MaxMessages int
	// MaxText bounds the rendered text of each record.
	MaxText uint32
}

// ParseMessages decodes count records that must occupy b exactly.
func ParseMessages(count uint16, b []byte, limits MessageLimits) ([]MessageRecord, error) {
	if int(count) > limits.MaxMessages {
		return nil, &hostserver.MalformedError{What: "messages", Offset: 22, Reason: fmt.Sprintf("%d records exceed the limit of %d", count, limits.MaxMessages)}
	}
	out := make([]MessageRecord, 0, count)
	off := 0
	for i := 0; i < int(count); i++ {
		rel := 24 + off // offset in the frame, for diagnostics
		if len(b)-off < MessageRecordMinLength {
			return nil, &hostserver.MalformedError{What: "message", Offset: rel, Reason: fmt.Sprintf("record %d: %d bytes remain, need %d", i, len(b)-off, MessageRecordMinLength)}
		}
		r := b[off:]
		ll := uint64(binary.BigEndian.Uint32(r))
		if ll < MessageRecordMinLength || ll > uint64(len(r)) {
			return nil, &hostserver.MalformedError{What: "message", Offset: rel, Reason: fmt.Sprintf("record %d length %d outside %d..%d", i, ll, MessageRecordMinLength, len(r))}
		}
		s := uint64(binary.BigEndian.Uint16(r[37:]))
		t := uint64(binary.BigEndian.Uint16(r[39:]))
		if MessageRecordMinLength+s+t > ll {
			return nil, &hostserver.MalformedError{What: "message", Offset: rel, Reason: fmt.Sprintf("record %d substitution %d + text %d exceed length %d", i, s, t, ll)}
		}
		if t > uint64(limits.MaxText) {
			return nil, &hostserver.MalformedError{What: "message", Offset: rel, Reason: fmt.Sprintf("record %d text of %d bytes exceeds the limit of %d", i, t, limits.MaxText)}
		}
		d0, d1 := r[13]&0x0F, r[14]&0x0F
		if d0 > 9 || d1 > 9 {
			return nil, &hostserver.MalformedError{What: "message", Offset: rel + 13, Reason: fmt.Sprintf("record %d message type digits are not decimal", i)}
		}
		rec := MessageRecord{
			Tag:          binary.BigEndian.Uint16(r[4:]),
			Type:         d0*10 + d1,
			Severity:     binary.BigEndian.Uint16(r[15:]),
			Substitution: append([]byte{}, r[41:41+s]...),
			Text:         append([]byte{}, r[41+s:41+s+t]...),
		}
		copy(rec.ID[:], r[6:13])
		copy(rec.File[:], r[17:27])
		copy(rec.Library[:], r[27:37])
		if end := 41 + s + t; end < ll {
			rec.Extension = append([]byte{}, r[end:ll]...)
		}
		out = append(out, rec)
		off += int(ll)
	}
	if off != len(b) {
		return nil, &hostserver.MalformedError{What: "messages", Offset: 24 + off, Reason: fmt.Sprintf("%d unexplained bytes after %d records", len(b)-off, count)}
	}
	return out, nil
}
