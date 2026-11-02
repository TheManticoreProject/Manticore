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
// 2.4.2.1 SID String Format): authorities below 2^32 are decimal, larger ones
// "0x" followed by exactly 12 lower-case hex digits, as RtlConvertSidToUnicodeString
// prints them.
func (s RPC_SID) String() string {
	auth := s.authority()
	var b strings.Builder
	b.WriteString("S-")
	b.WriteString(strconv.FormatUint(uint64(s.Revision), 10))
	b.WriteString("-")
	if auth >= 1<<32 {
		fmt.Fprintf(&b, "0x%012x", auth)
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

// sidMaxSubAuthorities is the largest SubAuthorityCount [MS-DTYP] 2.4.2.2 allows.
const sidMaxSubAuthorities = 15

// Marshal encodes the SID in its packed binary form ([MS-DTYP] 2.4.2.2): Revision,
// SubAuthorityCount, the 6-byte big-endian IdentifierAuthority, then each
// SubAuthority as a little-endian uint32. This is the form of the objectSid
// attribute and of SIDs inside security descriptors, not the NDR form used on
// RPC. SubAuthorityCount is taken from the length of SubAuthority.
//
// Returns:
// - The packed SID, 8 + 4*len(SubAuthority) bytes.
// - An error if there are more than 15 sub-authorities.
func (s RPC_SID) Marshal() ([]byte, error) {
	if len(s.SubAuthority) > sidMaxSubAuthorities {
		return nil, fmt.Errorf("dtyp: SID has %d sub-authorities, max %d", len(s.SubAuthority), sidMaxSubAuthorities)
	}
	out := make([]byte, 0, 8+4*len(s.SubAuthority))
	out = append(out, s.Revision, uint8(len(s.SubAuthority)))
	out = append(out, s.IdentifierAuthority[:]...)
	for _, sub := range s.SubAuthority {
		out = binary.LittleEndian.AppendUint32(out, sub)
	}
	return out, nil
}

// Unmarshal decodes a packed binary SID ([MS-DTYP] 2.4.2.2) from the start of data,
// as Marshal writes it.
//
// Returns:
// - The number of bytes consumed, 8 + 4*SubAuthorityCount.
// - An error if data is shorter than its SubAuthorityCount requires, or the count
//   exceeds 15.
func (s *RPC_SID) Unmarshal(data []byte) (int, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("dtyp: SID needs at least 8 bytes, got %d", len(data))
	}
	count := int(data[1])
	if count > sidMaxSubAuthorities {
		return 0, fmt.Errorf("dtyp: SID has %d sub-authorities, max %d", count, sidMaxSubAuthorities)
	}
	n := 8 + 4*count
	if len(data) < n {
		return 0, fmt.Errorf("dtyp: SID with %d sub-authorities needs %d bytes, got %d", count, n, len(data))
	}
	s.Revision = data[0]
	s.SubAuthorityCount = uint8(count)
	copy(s.IdentifierAuthority[:], data[2:8])
	s.SubAuthority = make([]uint32, count)
	for i := range s.SubAuthority {
		s.SubAuthority[i] = binary.LittleEndian.Uint32(data[8+4*i:])
	}
	return n, nil
}
