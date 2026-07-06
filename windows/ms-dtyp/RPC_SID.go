package msdtyp

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// RPC_SID is the [MS-DTYP] 2.4.2.3 marshallable security identifier. The SubAuthority
// member is a conformant array whose element count is given by SubAuthorityCount, so
// NDR hoists its maximum_count to the front of the structure ([C706] section 14.3.3.1);
// the walker derives both the hoisted count and SubAuthorityCount from the slice length,
// so callers set only SubAuthority (or use ParseSID).
//
// IdentifierAuthority is a 6-octet big-endian value transmitted verbatim ([MS-DTYP]
// 2.4.1 RPC_SID_IDENTIFIER_AUTHORITY). SubAuthorityCount is capped at 15 by the spec.
//
// Reference: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/5cb97814-a1c2-4215-b7dc-76d1f4bfad01
type RPC_SID struct {
	Revision            uint8
	SubAuthorityCount   uint8
	IdentifierAuthority [6]byte
	SubAuthority        []uint32 `ndr:"conformant,size_is=SubAuthorityCount"`
}

// authority returns the 48-bit identifier authority as a single value.
func (s RPC_SID) authority() uint64 {
	var b [8]byte
	copy(b[2:], s.IdentifierAuthority[:])
	return binary.BigEndian.Uint64(b[:])
}

// String renders the SID in the standard "S-R-I-S1-S2-…" textual form ([MS-DTYP]
// 2.4.2.1 SID String Format): authorities below 2^32 are decimal, larger ones hex.
func (s RPC_SID) String() string {
	auth := s.authority()
	var b strings.Builder
	b.WriteString("S-")
	b.WriteString(strconv.FormatUint(uint64(s.Revision), 10))
	b.WriteString("-")
	if auth >= 1<<32 {
		b.WriteString("0x")
		b.WriteString(strconv.FormatUint(auth, 16))
	} else {
		b.WriteString(strconv.FormatUint(auth, 10))
	}
	for _, sub := range s.SubAuthority {
		b.WriteString("-")
		b.WriteString(strconv.FormatUint(uint64(sub), 10))
	}
	return b.String()
}

// ParseSID parses a textual SID ("S-1-5-21-…") into an RPC_SID, setting Revision,
// IdentifierAuthority, and the SubAuthority array (SubAuthorityCount is derived on
// marshal). The identifier authority may be decimal or 0x-prefixed hexadecimal.
func ParseSID(s string) (RPC_SID, error) {
	parts := strings.Split(s, "-")
	if len(parts) < 3 || parts[0] != "S" {
		return RPC_SID{}, fmt.Errorf("dtyp: invalid SID %q", s)
	}
	rev, err := strconv.ParseUint(parts[1], 10, 8)
	if err != nil {
		return RPC_SID{}, fmt.Errorf("dtyp: invalid SID revision in %q: %w", s, err)
	}
	var auth uint64
	if strings.HasPrefix(parts[2], "0x") || strings.HasPrefix(parts[2], "0X") {
		auth, err = strconv.ParseUint(parts[2][2:], 16, 48)
	} else {
		auth, err = strconv.ParseUint(parts[2], 10, 48)
	}
	if err != nil {
		return RPC_SID{}, fmt.Errorf("dtyp: invalid SID authority in %q: %w", s, err)
	}
	subs := parts[3:]
	if len(subs) > 15 {
		return RPC_SID{}, fmt.Errorf("dtyp: SID %q has %d sub-authorities, max 15", s, len(subs))
	}
	sid := RPC_SID{Revision: uint8(rev), SubAuthorityCount: uint8(len(subs))}
	var ab [8]byte
	binary.BigEndian.PutUint64(ab[:], auth)
	copy(sid.IdentifierAuthority[:], ab[2:])
	for _, p := range subs {
		v, err := strconv.ParseUint(p, 10, 32)
		if err != nil {
			return RPC_SID{}, fmt.Errorf("dtyp: invalid SID sub-authority %q in %q: %w", p, s, err)
		}
		sid.SubAuthority = append(sid.SubAuthority, uint32(v))
	}
	return sid, nil
}
