package message

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// MessageRecordMinLength is the fixed part of a classic message record.
const MessageRecordMinLength = 41

// Message record format tags (the 2 bytes after a record's length). The
// message option byte selects the format: options 0 and 2 return classic
// 0x1102 records, the level-10 options 3 and 4 return 0x1106 and the
// level-11 options 5 and 6 return 0x1107 (observed on V7R5M0, 2026-10-19).
// Any other tag is parsed as a classic record, as before.
const (
	TagClassic  uint16 = 0x1102
	TagExtended uint16 = 0x1106
	TagRCVM0300 uint16 = 0x1107
)

// MessageRecord is one IBM message record with its character fields still
// encoded: in the job CCSID, or in TextCCSID when that is set.
type MessageRecord struct {
	// Tag is the 2-byte format tag.
	Tag      uint16
	ID       [7]byte
	Type     uint8
	Severity uint16
	File     [10]byte
	Library  [10]byte
	// Substitution and Text are copies, safe to retain.
	Substitution []byte
	Text         []byte
	// Help is the second-level text (0x1106 and 0x1107 only).
	Help []byte
	// TextCCSID and DataCCSID tag Text/Help and Substitution (0x1106 and
	// 0x1107 only); 0 or 65535 means the job CCSID.
	TextCCSID, DataCCSID uint32
	// Extension holds record bytes after the last field this package
	// decodes, if any.
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
		if tag := binary.BigEndian.Uint16(r[4:]); tag == TagExtended || tag == TagRCVM0300 {
			rec, err := parseFieldRecord(tag, r[6:ll], limits)
			if err != nil {
				return nil, &hostserver.MalformedError{What: "message", Offset: rel, Reason: fmt.Sprintf("record %d (0x%04X): %v", i, tag, err)}
			}
			out = append(out, rec)
			off += int(ll)
			continue
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

// fieldReader walks the 4-byte-length-prefixed fields of 0x1106 and
// 0x1107 records.
type fieldReader struct {
	b   []byte
	err error
}

func (f *fieldReader) next() []byte {
	if f.err != nil {
		return nil
	}
	if len(f.b) < 4 {
		f.err = fmt.Errorf("%d bytes cannot hold a field length", len(f.b))
		return nil
	}
	n := uint64(binary.BigEndian.Uint32(f.b))
	if n > uint64(len(f.b)-4) {
		f.err = fmt.Errorf("field length %d exceeds the %d remaining bytes", n, len(f.b)-4)
		return nil
	}
	v := f.b[4 : 4+n]
	f.b = f.b[4+n:]
	return v
}

func (f *fieldReader) fixed(n int) []byte {
	if f.err != nil {
		return nil
	}
	if len(f.b) < n {
		f.err = fmt.Errorf("%d bytes cannot hold a %d-byte value", len(f.b), n)
		return nil
	}
	v := f.b[:n]
	f.b = f.b[n:]
	return v
}

func (f *fieldReader) integer(width int) uint64 {
	v := f.next()
	if f.err == nil && len(v) != width {
		f.err = fmt.Errorf("integer field of %d bytes, want %d", len(v), width)
	}
	if f.err != nil {
		return 0
	}
	var x uint64
	for _, c := range v {
		x = x<<8 | uint64(c)
	}
	return x
}

// padded copies v into a blank-padded fixed field.
func padded(dst []byte, v []byte, f *fieldReader, what string) {
	if f.err == nil && len(v) > len(dst) {
		f.err = fmt.Errorf("%s of %d bytes exceeds %d", what, len(v), len(dst))
		return
	}
	for i := range dst {
		dst[i] = 0x40
	}
	copy(dst, v)
}

// messageType decodes the two EBCDIC digits of a message type.
func messageType(v []byte, f *fieldReader) uint8 {
	if f.err != nil {
		return 0
	}
	if len(v) != 2 || v[0]&0xF0 != 0xF0 || v[1]&0xF0 != 0xF0 || v[0]&0x0F > 9 || v[1]&0x0F > 9 {
		f.err = fmt.Errorf("message type is not two EBCDIC digits")
		return 0
	}
	return (v[0]&0x0F)*10 + v[1]&0x0F
}

// parseFieldRecord decodes the body (after LL and tag) of a 0x1106 or
// 0x1107 record.
//
// 0x1106: text CCSID (4), data CCSID (4), severity (2), then length-prefixed
// type, ID, message file, library, text, substitution data and help.
//
// 0x1107: 26 length-prefixed fields in the order of IBM's QMHRCVPM format
// RCVM0300 (severity, ID, type, key, message file, library specified,
// library used, sending job, user and number, sending program and
// instruction, date, time, receiving program and instruction, sending and
// receiving type, text and data CCSID conversion status, alert option, text
// CCSID, data CCSID, substitution data, text, help). Fields this package
// does not expose are skipped.
func parseFieldRecord(tag uint16, b []byte, limits MessageLimits) (MessageRecord, error) {
	f := &fieldReader{b: b}
	rec := MessageRecord{Tag: tag}
	var id, file, lib, text, subst, help []byte
	if tag == TagExtended {
		rec.TextCCSID = uint32(binaryUint(f.fixed(4)))
		rec.DataCCSID = uint32(binaryUint(f.fixed(4)))
		rec.Severity = uint16(binaryUint(f.fixed(2)))
		rec.Type = messageType(f.next(), f)
		id, file, lib = f.next(), f.next(), f.next()
		text, subst, help = f.next(), f.next(), f.next()
	} else {
		sev := f.integer(4)
		if f.err == nil && sev > 99 {
			f.err = fmt.Errorf("severity %d above 99", sev)
		}
		rec.Severity = uint16(sev)
		id = f.next()
		rec.Type = messageType(f.next(), f)
		f.next() // key
		file = f.next()
		f.next() // library specified
		lib = f.next()
		for i := 0; i < 13; i++ {
			// sending job, user, number, program, instruction, date, time,
			// receiving program, instruction, sending type, receiving
			// type, text and data CCSID conversion status
			f.next()
		}
		f.next() // alert option
		rec.TextCCSID = uint32(f.integer(4))
		rec.DataCCSID = uint32(f.integer(4))
		subst, text, help = f.next(), f.next(), f.next()
	}
	padded(rec.ID[:], id, f, "message ID")
	padded(rec.File[:], file, f, "message file")
	padded(rec.Library[:], lib, f, "library")
	if f.err == nil && (uint64(len(text)) > uint64(limits.MaxText) || uint64(len(help)) > uint64(limits.MaxText)) {
		f.err = fmt.Errorf("text of %d or help of %d bytes exceeds the limit of %d", len(text), len(help), limits.MaxText)
	}
	if f.err != nil {
		return MessageRecord{}, f.err
	}
	rec.Text = append([]byte{}, text...)
	rec.Substitution = append([]byte{}, subst...)
	rec.Help = append([]byte{}, help...)
	if len(f.b) > 0 {
		rec.Extension = append([]byte{}, f.b...)
	}
	return rec, nil
}

func binaryUint(v []byte) uint64 {
	var x uint64
	for _, c := range v {
		x = x<<8 | uint64(c)
	}
	return x
}
