package msdrsr

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
)

// Well-known attribute OIDs used by DCSync ([MS-ADA1]/[MS-ADA3]). GetNCChanges keys each
// attribute by an ATTRTYP that compresses one of these OIDs against the reply's prefix
// table; attidForOID resolves an OID to its ATTRTYP using that table.
const (
	oidObjectSid               = "1.2.840.113556.1.4.146"
	oidSAMAccountName          = "1.2.840.113556.1.4.221"
	oidUnicodePwd              = "1.2.840.113556.1.4.90"
	oidDBCSPwd                 = "1.2.840.113556.1.4.55"
	oidNTPwdHistory            = "1.2.840.113556.1.4.94"
	oidLMPwdHistory            = "1.2.840.113556.1.4.160"
	oidSupplementalCredentials = "1.2.840.113556.1.4.125"
	oidUserPrincipalName       = "1.2.840.113556.1.4.656"
)

// berEncodeOIDContent returns the BER/DER content octets of a dotted OID (without the
// 0x06 tag and length prefix): the first two arcs combine as 40*a+b, and every arc is
// base-128 big-endian with the continuation bit set on all but its last octet
// ([ITU-T X.690] 8.19).
func berEncodeOIDContent(oid string) ([]byte, error) {
	parts := strings.Split(oid, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("drsuapi: invalid OID %q", oid)
	}
	arcs := make([]uint64, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("drsuapi: invalid OID arc %q: %w", p, err)
		}
		arcs[i] = v
	}
	var out []byte
	out = append(out, base128(40*arcs[0]+arcs[1])...)
	for _, a := range arcs[2:] {
		out = append(out, base128(a)...)
	}
	return out, nil
}

// base128 encodes v as big-endian base-128 with the high (continuation) bit set on every
// octet except the last.
func base128(v uint64) []byte {
	if v == 0 {
		return []byte{0}
	}
	var rev []byte
	for v > 0 {
		rev = append(rev, byte(v&0x7f))
		v >>= 7
	}
	out := make([]byte, len(rev))
	for i := range rev {
		out[i] = rev[len(rev)-1-i]
	}
	for i := 0; i < len(out)-1; i++ {
		out[i] |= 0x80
	}
	return out
}

// attidForOID computes the ATTRTYP that the server used for the given OID, by matching
// the OID's prefix (its BER content minus the last arc's octet(s)) against the reply's
// prefix table and combining the entry's Ndx with the last arc ([MS-DRSR] 5.16.4,
// MakeAttid). It returns false if the prefix is not present in the table (the attribute
// was not part of the reply).
func attidForOID(prefixTable structures.SCHEMA_PREFIX_TABLE, oid string) (uint32, bool) {
	parts := strings.Split(oid, ".")
	lastValue, err := strconv.ParseUint(parts[len(parts)-1], 10, 32)
	if err != nil {
		return 0, false
	}
	content, err := berEncodeOIDContent(oid)
	if err != nil {
		return 0, false
	}
	var prefix []byte
	if lastValue < 128 {
		prefix = content[:len(content)-1]
	} else {
		prefix = content[:len(content)-2]
	}

	for _, e := range prefixTable.PPrefixEntry {
		elems := e.Prefix.Elements
		if int(e.Prefix.Length) <= len(elems) {
			elems = elems[:e.Prefix.Length]
		}
		if !bytes.Equal(elems, prefix) {
			continue
		}
		lowerWord := uint32(lastValue % 16384)
		if lastValue >= 16384 {
			lowerWord += 32768
		}
		return (uint32(e.Ndx) << 16) | lowerWord, true
	}
	return 0, false
}

// findAttr returns the values of the attribute with the given OID in obj, resolved
// through prefixTable, or nil if the object does not carry it.
func findAttr(obj ReplicatedObject, prefixTable structures.SCHEMA_PREFIX_TABLE, oid string) [][]byte {
	attid, ok := attidForOID(prefixTable, oid)
	if !ok {
		return nil
	}
	for _, a := range obj.Attributes {
		if a.AttrType == attid {
			return a.Values
		}
	}
	return nil
}
