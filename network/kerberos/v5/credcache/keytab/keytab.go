// Package keytab reads and writes the MIT/Heimdal Kerberos keytab file
// (the ".keytab" / krb5.keytab format) in versioned format 2 (magic 0x05 0x02),
// the interchange format Unix Kerberos tooling (ktutil, kinit -kt, klist -k) uses
// to store a principal's long-term keys on disk.
//
// A keytab holds one or more entries, each binding a principal to a single
// long-term key (an enctype plus the raw key bytes) and a key version number
// (kvno). Unlike a credential cache (see the ccache package) it stores permanent
// keys, not tickets, so it lets a client authenticate non-interactively — the
// keytab-based analogue of pass-the-key.
//
// Wire format (documented at
// https://web.mit.edu/kerberos/krb5-devel/doc/formats/keytab_file_format.html):
//
//	keytab {
//	    uint16                 file_format_version = 0x0502
//	    keytab_entry           entries[*]         (runs to EOF)
//	}
//	keytab_entry {
//	    int32                  size               // bytes that FOLLOW this field;
//	                                              // negative = a hole (deleted entry),
//	                                              // its magnitude is the hole length
//	    uint16                 num_components     // v2 does NOT count the realm
//	    counted_octet_string   realm
//	    counted_octet_string   components[num_components]
//	    uint32                 name_type          // v2 only
//	    uint32                 timestamp          // Unix seconds
//	    uint8                  vno8               // 8-bit key version number
//	    keyblock               key
//	    uint32                 vno                // OPTIONAL: present when >= 4 bytes
//	                                              // remain in size and it is non-zero;
//	                                              // supersedes vno8
//	}
//	keyblock             { uint16 enctype; counted_octet_string key }
//	counted_octet_string { uint16 length; uint8 data[length] }
//
// Version 2 is big-endian throughout (version 1 used host byte order and folded
// the realm into num_components; only version 2 is emitted today and is what this
// package writes). This package implements the wire format and entry selection;
// wiring a selected key into a live client is done by the kerberos client layer.
package keytab

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// Version2 is the file_format_version this package reads and writes: the first
// byte is always 0x05, the second is the version (2).
const Version2 = 0x0502

// Principal identifies the account a key belongs to: one or more name
// components, a realm, and the RFC 4120 name-type.
type Principal struct {
	NameType   uint32
	Realm      string
	Components []string
}

// String renders the principal in the conventional "comp1/comp2@REALM" form.
func (p Principal) String() string {
	return strings.Join(p.Components, "/") + "@" + p.Realm
}

// Entry is a single keytab record: one long-term key for one principal.
type Entry struct {
	Principal Principal
	// Timestamp is when the entry was written, in Unix seconds (0 if unknown).
	Timestamp uint32
	// KVNO8 is the 8-bit key version number always present on the wire.
	KVNO8 uint8
	// KVNO is the optional 32-bit key version number. When non-zero it is written
	// after the key and supersedes KVNO8 (kvnos above 255 need it). Zero means the
	// 32-bit field is absent and KVNO8 is authoritative.
	KVNO uint32
	// EType is the key's encryption type (see iana.EType*).
	EType uint16
	// Key is the raw long-term key bytes.
	Key []byte
}

// Kvno returns the effective key version number: the 32-bit KVNO when non-zero,
// otherwise the 8-bit KVNO8.
func (e Entry) Kvno() uint32 {
	if e.KVNO != 0 {
		return e.KVNO
	}
	return uint32(e.KVNO8)
}

// Keytab is a parsed keytab: an ordered list of key entries.
type Keytab struct {
	Entries []Entry
}

// New returns an empty keytab.
func New() *Keytab { return &Keytab{} }

// Add appends an entry binding a key to a principal. The 8-bit KVNO8 is set from
// kvno (its low byte) and, when kvno exceeds 255, the 32-bit KVNO too. NameType
// defaults to NT-PRINCIPAL when zero. The key bytes are copied.
func (kt *Keytab) Add(principal Principal, etype int, key []byte, kvno uint32) {
	if principal.NameType == 0 {
		principal.NameType = iana.NameTypePrincipal
	}
	k := make([]byte, len(key))
	copy(k, key)
	e := Entry{
		Principal: principal,
		KVNO8:     uint8(kvno),
		EType:     uint16(etype),
		Key:       k,
	}
	if kvno > 0xff {
		e.KVNO = kvno
	}
	kt.Entries = append(kt.Entries, e)
}

// ── selection ─────────────────────────────────────────────────────────────

// etypeStrength orders enctypes strongest-first for "best entry" selection.
func etypeStrength(etype uint16) int {
	switch int(etype) {
	case iana.ETypeAES256CTSHMACSHA196:
		return 5
	case iana.ETypeAES256CTSHMACSHA384:
		return 4
	case iana.ETypeAES128CTSHMACSHA196:
		return 3
	case iana.ETypeAES128CTSHMACSHA256:
		return 2
	case iana.ETypeRC4HMAC:
		return 1
	default:
		return 0
	}
}

// principalMatches reports whether entry principal p matches the query string,
// case-insensitively on the realm. The query is "comp1/comp2@REALM"; a missing
// "@REALM" matches any realm, and an empty query matches anything.
func principalMatches(p Principal, query string) bool {
	if query == "" {
		return true
	}
	name := query
	realm := ""
	if at := strings.LastIndexByte(query, '@'); at >= 0 {
		name = query[:at]
		realm = query[at+1:]
	}
	if realm != "" && !strings.EqualFold(realm, p.Realm) {
		return false
	}
	return name == strings.Join(p.Components, "/")
}

// Find returns pointers to every entry matching the filter, in file order.
// A principal of "" matches any principal; etype <= 0 matches any enctype; and
// kvno < 0 matches any key version.
func (kt *Keytab) Find(principal string, etype int, kvno int) []*Entry {
	var out []*Entry
	for i := range kt.Entries {
		e := &kt.Entries[i]
		if !principalMatches(e.Principal, principal) {
			continue
		}
		if etype > 0 && int(e.EType) != etype {
			continue
		}
		if kvno >= 0 && int(e.Kvno()) != kvno {
			continue
		}
		out = append(out, e)
	}
	return out
}

// Select returns the single best entry matching the filter, or nil when none
// match. "Best" is the highest key version number, breaking ties by enctype
// strength (AES256 > AES256-SHA2 > AES128 > AES128-SHA2 > RC4 > other). Pass
// principal "" for any principal, etype <= 0 for any enctype, and kvno < 0 to
// select the newest key rather than a specific version.
func (kt *Keytab) Select(principal string, etype int, kvno int) *Entry {
	var best *Entry
	for _, e := range kt.Find(principal, etype, kvno) {
		if best == nil {
			best = e
			continue
		}
		if e.Kvno() > best.Kvno() ||
			(e.Kvno() == best.Kvno() && etypeStrength(e.EType) > etypeStrength(best.EType)) {
			best = e
		}
	}
	return best
}

// ── marshaling ──────────────────────────────────────────────────────────────

func putOctetString(buf *bytes.Buffer, data []byte) {
	var l [2]byte
	binary.BigEndian.PutUint16(l[:], uint16(len(data)))
	buf.Write(l[:])
	buf.Write(data)
}

func putU16(buf *bytes.Buffer, v uint16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], v)
	buf.Write(b[:])
}

func putU32(buf *bytes.Buffer, v uint32) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	buf.Write(b[:])
}

// marshalEntryBody encodes everything that the entry's size prefix counts: the
// principal, timestamp, vno8, keyblock and optional 32-bit kvno.
func marshalEntryBody(e Entry) []byte {
	var b bytes.Buffer
	// v2 num_components excludes the realm.
	putU16(&b, uint16(len(e.Principal.Components)))
	putOctetString(&b, []byte(e.Principal.Realm))
	for _, c := range e.Principal.Components {
		putOctetString(&b, []byte(c))
	}
	nt := e.Principal.NameType
	if nt == 0 {
		nt = iana.NameTypePrincipal
	}
	putU32(&b, nt)
	putU32(&b, e.Timestamp)
	b.WriteByte(e.KVNO8)
	// keyblock: enctype + counted key.
	putU16(&b, e.EType)
	putOctetString(&b, e.Key)
	// Optional 32-bit kvno, written only when set (supersedes vno8).
	if e.KVNO != 0 {
		putU32(&b, e.KVNO)
	}
	return b.Bytes()
}

// Marshal encodes the keytab in versioned format 2 (0x0502, big-endian).
func (kt *Keytab) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	putU16(&buf, Version2)
	for i := range kt.Entries {
		body := marshalEntryBody(kt.Entries[i])
		// The size prefix counts the bytes that follow it (the whole entry body),
		// not itself. It is written signed; a positive value is a valid entry.
		putU32(&buf, uint32(int32(len(body))))
		buf.Write(body)
	}
	return buf.Bytes(), nil
}

// ── unmarshaling ────────────────────────────────────────────────────────────

// reader is a bounds-checked big-endian cursor over the keytab bytes.
type reader struct {
	b   []byte
	pos int
}

func (r *reader) need(n int) error {
	if n < 0 || r.pos+n > len(r.b) {
		return fmt.Errorf("keytab: truncated: need %d bytes at offset %d of %d", n, r.pos, len(r.b))
	}
	return nil
}

func (r *reader) u16() (uint16, error) {
	if err := r.need(2); err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint16(r.b[r.pos:])
	r.pos += 2
	return v, nil
}

func (r *reader) i32() (int32, error) {
	if err := r.need(4); err != nil {
		return 0, err
	}
	v := int32(binary.BigEndian.Uint32(r.b[r.pos:]))
	r.pos += 4
	return v, nil
}

// octetString reads a 16-bit-length-prefixed byte string.
func (r *reader) octetString() ([]byte, error) {
	n, err := r.u16()
	if err != nil {
		return nil, err
	}
	if err := r.need(int(n)); err != nil {
		return nil, err
	}
	out := make([]byte, n)
	copy(out, r.b[r.pos:r.pos+int(n)])
	r.pos += int(n)
	return out, nil
}

// parseEntryBody decodes one entry from exactly body (the size-prefixed slice).
func parseEntryBody(body []byte) (Entry, error) {
	var e Entry
	r := &reader{b: body}

	nComp, err := r.u16()
	if err != nil {
		return e, err
	}
	realm, err := r.octetString()
	if err != nil {
		return e, err
	}
	e.Principal.Realm = string(realm)
	for i := uint16(0); i < nComp; i++ {
		c, err := r.octetString()
		if err != nil {
			return e, err
		}
		e.Principal.Components = append(e.Principal.Components, string(c))
	}
	nt, err := r.i32()
	if err != nil {
		return e, err
	}
	e.Principal.NameType = uint32(nt)
	ts, err := r.i32()
	if err != nil {
		return e, err
	}
	e.Timestamp = uint32(ts)
	if err := r.need(1); err != nil {
		return e, err
	}
	e.KVNO8 = r.b[r.pos]
	r.pos++
	if e.EType, err = r.u16(); err != nil {
		return e, err
	}
	if e.Key, err = r.octetString(); err != nil {
		return e, err
	}
	// Optional 32-bit kvno: present only if >= 4 bytes remain and it is non-zero.
	if len(body)-r.pos >= 4 {
		v, err := r.i32()
		if err != nil {
			return e, err
		}
		if v != 0 {
			e.KVNO = uint32(v)
		}
	}
	return e, nil
}

// Unmarshal parses a keytab. Only versioned format 2 (0x0502) is supported.
func Unmarshal(data []byte) (*Keytab, error) {
	r := &reader{b: data}
	ver, err := r.u16()
	if err != nil {
		return nil, err
	}
	if ver != Version2 {
		return nil, fmt.Errorf("keytab: unsupported version 0x%04x (only 0x%04x supported)", ver, Version2)
	}

	kt := &Keytab{}
	for r.pos < len(data) {
		size, err := r.i32()
		if err != nil {
			return nil, err
		}
		// A length of 0 ends the file; a negative length is a hole (deleted
		// entry) whose magnitude is the number of bytes to skip.
		if size == 0 {
			break
		}
		if size < 0 {
			skip := int(-int64(size))
			if err := r.need(skip); err != nil {
				return nil, err
			}
			r.pos += skip
			continue
		}
		if err := r.need(int(size)); err != nil {
			return nil, err
		}
		body := r.b[r.pos : r.pos+int(size)]
		r.pos += int(size)
		e, err := parseEntryBody(body)
		if err != nil {
			return nil, err
		}
		kt.Entries = append(kt.Entries, e)
	}
	return kt, nil
}
