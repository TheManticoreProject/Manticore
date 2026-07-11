// Package ccache reads and writes the MIT Kerberos FILE credential cache
// (KRB5CCNAME) in format version 4, the interchange format used by Linux
// Kerberos tooling. Version 4 is big-endian throughout and is
// documented at
// https://web.mit.edu/kerberos/krb5-latest/doc/formats/ccache_file_format.html.
//
// Layout: a two-byte version tag (0x05 0x04), a tagged header, the default
// client principal, then a sequence of credential entries running to EOF (no
// count, no terminator). This package implements the wire format; building
// entries from a live TGT/ST is done by the client layer.
package ccache

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Version4 is the file_format_version this package reads and writes.
const Version4 = 0x0504

// headerTagDeltaTime is the only defined v4 header field: two uint32s (seconds,
// microseconds) giving the KDC-relative time offset.
const headerTagDeltaTime = 1

// Principal is a cache principal (client or server).
type Principal struct {
	NameType   uint32
	Realm      string
	Components []string
}

// Keyblock is a session key with its encryption type.
type Keyblock struct {
	EType    uint16
	KeyValue []byte
}

// Address is a host address entry.
type Address struct {
	AddrType uint16
	Data     []byte
}

// AuthData is an authorization-data entry.
type AuthData struct {
	ADType uint16
	Data   []byte
}

// Credential is one cached ticket. Times are Unix seconds (uint32) as stored on
// the wire; TicketFlags uses the RFC 4120 ticket-flag bit layout (bit 0 = MSB).
type Credential struct {
	Client       Principal
	Server       Principal
	Key          Keyblock
	AuthTime     uint32
	StartTime    uint32
	EndTime      uint32
	RenewTill    uint32
	IsSKey       bool
	TicketFlags  uint32
	Addresses    []Address
	AuthData     []AuthData
	Ticket       []byte
	SecondTicket []byte
}

// CCache is a parsed credential cache.
type CCache struct {
	// DeltaTime is the optional KDC time offset (seconds, microseconds). Zero
	// means no delta-time header field is written.
	DeltaTimeSecs    uint32
	DeltaTimeUsecs   uint32
	DefaultPrincipal Principal
	Credentials      []Credential
}

// ── marshaling ──────────────────────────────────────────────────────────────

func putData(buf *bytes.Buffer, data []byte) {
	var l [4]byte
	binary.BigEndian.PutUint32(l[:], uint32(len(data)))
	buf.Write(l[:])
	buf.Write(data)
}

func putPrincipal(buf *bytes.Buffer, p Principal) {
	var u [4]byte
	binary.BigEndian.PutUint32(u[:], p.NameType)
	buf.Write(u[:])
	// count of components does NOT include the realm in v2+.
	binary.BigEndian.PutUint32(u[:], uint32(len(p.Components)))
	buf.Write(u[:])
	putData(buf, []byte(p.Realm))
	for _, c := range p.Components {
		putData(buf, []byte(c))
	}
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

func putCredential(buf *bytes.Buffer, c Credential) {
	putPrincipal(buf, c.Client)
	putPrincipal(buf, c.Server)
	// keyblock: v4 has a single enctype field followed by the key data.
	putU16(buf, c.Key.EType)
	putData(buf, c.Key.KeyValue)
	putU32(buf, c.AuthTime)
	putU32(buf, c.StartTime)
	putU32(buf, c.EndTime)
	putU32(buf, c.RenewTill)
	if c.IsSKey {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	putU32(buf, c.TicketFlags)
	putU32(buf, uint32(len(c.Addresses)))
	for _, a := range c.Addresses {
		putU16(buf, a.AddrType)
		putData(buf, a.Data)
	}
	putU32(buf, uint32(len(c.AuthData)))
	for _, a := range c.AuthData {
		putU16(buf, a.ADType)
		putData(buf, a.Data)
	}
	putData(buf, c.Ticket)
	putData(buf, c.SecondTicket)
}

// Marshal encodes the cache in ccache format version 4.
func (cc *CCache) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	putU16(&buf, Version4)

	// Header: only the delta-time field, and only if set.
	var hdr bytes.Buffer
	if cc.DeltaTimeSecs != 0 || cc.DeltaTimeUsecs != 0 {
		putU16(&hdr, headerTagDeltaTime)
		putU16(&hdr, 8)
		putU32(&hdr, cc.DeltaTimeSecs)
		putU32(&hdr, cc.DeltaTimeUsecs)
	}
	putU16(&buf, uint16(hdr.Len()))
	buf.Write(hdr.Bytes())

	putPrincipal(&buf, cc.DefaultPrincipal)
	for i := range cc.Credentials {
		putCredential(&buf, cc.Credentials[i])
	}
	return buf.Bytes(), nil
}

// ── unmarshaling ──────────────────────────────────────────────────────────────

// reader is a bounds-checked big-endian cursor over the cache bytes.
type reader struct {
	b   []byte
	pos int
}

func (r *reader) need(n int) error {
	if n < 0 || r.pos+n > len(r.b) {
		return fmt.Errorf("ccache: truncated: need %d bytes at offset %d of %d", n, r.pos, len(r.b))
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

func (r *reader) u32() (uint32, error) {
	if err := r.need(4); err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint32(r.b[r.pos:])
	r.pos += 4
	return v, nil
}

func (r *reader) u8() (byte, error) {
	if err := r.need(1); err != nil {
		return 0, err
	}
	v := r.b[r.pos]
	r.pos++
	return v, nil
}

func (r *reader) data() ([]byte, error) {
	n, err := r.u32()
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

func (r *reader) principal() (Principal, error) {
	var p Principal
	nt, err := r.u32()
	if err != nil {
		return p, err
	}
	p.NameType = nt
	n, err := r.u32()
	if err != nil {
		return p, err
	}
	realm, err := r.data()
	if err != nil {
		return p, err
	}
	p.Realm = string(realm)
	for i := uint32(0); i < n; i++ {
		c, err := r.data()
		if err != nil {
			return p, err
		}
		p.Components = append(p.Components, string(c))
	}
	return p, nil
}

func (r *reader) credential() (Credential, error) {
	var c Credential
	var err error
	if c.Client, err = r.principal(); err != nil {
		return c, err
	}
	if c.Server, err = r.principal(); err != nil {
		return c, err
	}
	if c.Key.EType, err = r.u16(); err != nil {
		return c, err
	}
	if c.Key.KeyValue, err = r.data(); err != nil {
		return c, err
	}
	if c.AuthTime, err = r.u32(); err != nil {
		return c, err
	}
	if c.StartTime, err = r.u32(); err != nil {
		return c, err
	}
	if c.EndTime, err = r.u32(); err != nil {
		return c, err
	}
	if c.RenewTill, err = r.u32(); err != nil {
		return c, err
	}
	skey, err := r.u8()
	if err != nil {
		return c, err
	}
	c.IsSKey = skey != 0
	if c.TicketFlags, err = r.u32(); err != nil {
		return c, err
	}
	nAddr, err := r.u32()
	if err != nil {
		return c, err
	}
	for i := uint32(0); i < nAddr; i++ {
		at, err := r.u16()
		if err != nil {
			return c, err
		}
		d, err := r.data()
		if err != nil {
			return c, err
		}
		c.Addresses = append(c.Addresses, Address{AddrType: at, Data: d})
	}
	nAD, err := r.u32()
	if err != nil {
		return c, err
	}
	for i := uint32(0); i < nAD; i++ {
		adt, err := r.u16()
		if err != nil {
			return c, err
		}
		d, err := r.data()
		if err != nil {
			return c, err
		}
		c.AuthData = append(c.AuthData, AuthData{ADType: adt, Data: d})
	}
	if c.Ticket, err = r.data(); err != nil {
		return c, err
	}
	if c.SecondTicket, err = r.data(); err != nil {
		return c, err
	}
	return c, nil
}

// Unmarshal parses a ccache. Only format version 4 is supported.
func Unmarshal(data []byte) (*CCache, error) {
	r := &reader{b: data}
	ver, err := r.u16()
	if err != nil {
		return nil, err
	}
	if ver != Version4 {
		return nil, fmt.Errorf("ccache: unsupported version 0x%04x (only 0x%04x/v4 supported)", ver, Version4)
	}

	cc := &CCache{}
	hdrLen, err := r.u16()
	if err != nil {
		return nil, err
	}
	if err := r.need(int(hdrLen)); err != nil {
		return nil, err
	}
	hdrEnd := r.pos + int(hdrLen)
	for r.pos < hdrEnd {
		tag, err := r.u16()
		if err != nil {
			return nil, err
		}
		flen, err := r.u16()
		if err != nil {
			return nil, err
		}
		if err := r.need(int(flen)); err != nil {
			return nil, err
		}
		if tag == headerTagDeltaTime && flen == 8 {
			cc.DeltaTimeSecs = binary.BigEndian.Uint32(r.b[r.pos:])
			cc.DeltaTimeUsecs = binary.BigEndian.Uint32(r.b[r.pos+4:])
		}
		r.pos += int(flen) // skip unknown/known field body
	}
	r.pos = hdrEnd

	if cc.DefaultPrincipal, err = r.principal(); err != nil {
		return nil, err
	}
	for r.pos < len(data) {
		cred, err := r.credential()
		if err != nil {
			return nil, err
		}
		cc.Credentials = append(cc.Credentials, cred)
	}
	return cc, nil
}
