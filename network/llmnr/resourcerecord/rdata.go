package resourcerecord

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/llmnr/domain_name"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
)

// This file implements typed encode/decode of the RDATA field for the resource
// record types commonly seen over LLMNR/DNS: A, AAAA, PTR, CNAME, NS, TXT and
// SRV. The wire formats follow RFC 1035 §3.3 (A/CNAME/NS/PTR/TXT) and RFC 2782
// (SRV). RDATA for unmodeled types is left as the opaque RData byte slice.
//
// Name-bearing RDATA (PTR/CNAME/NS and the SRV target) may contain 0xC0
// compression pointers whose 14-bit offset is measured from the start of the
// enclosing message (RFC 1035 §4.1.4). The typed decoders below therefore
// resolve names through the message context captured by UnmarshalFromMessage
// (rr.message / rr.rdataOffset) rather than against the RData sub-slice, which
// would resolve those pointers against the wrong origin.

// decodeName decodes a single domain name that begins at rdataInternalOffset
// bytes into this record's RData. When the record was decoded from a full
// message (via UnmarshalFromMessage) the name is resolved against the message
// origin so that compression pointers work; otherwise it falls back to decoding
// from the RData bytes alone, which only succeeds for uncompressed names.
func (rr *ResourceRecord) decodeName(rdataInternalOffset int) (string, error) {
	if rdataInternalOffset > len(rr.RData) {
		return "", fmt.Errorf("rdata too short for name at offset %d", rdataInternalOffset)
	}

	if rr.message != nil {
		name, _, err := domain_name.DecodeDomainName(rr.message, rr.rdataOffset+rdataInternalOffset)
		if err != nil {
			return "", err
		}
		return name, nil
	}

	name, _, err := domain_name.DecodeDomainName(rr.RData, rdataInternalOffset)
	if err != nil {
		return "", err
	}
	return name, nil
}

// AsA returns the IPv4 address carried by an A record. It errors if the record
// is not an A record or if RDATA is not exactly four octets (RFC 1035 §3.4.1).
func (rr *ResourceRecord) AsA() (net.IP, error) {
	if rr.Type != llmnr_type.TypeA {
		return nil, fmt.Errorf("record type is %s, not A", rr.Type.String())
	}
	if len(rr.RData) != net.IPv4len {
		return nil, fmt.Errorf("A rdata length is %d, want %d", len(rr.RData), net.IPv4len)
	}
	ip := make(net.IP, net.IPv4len)
	copy(ip, rr.RData)
	return ip, nil
}

// AsAAAA returns the IPv6 address carried by an AAAA record. It errors if the
// record is not an AAAA record or if RDATA is not exactly sixteen octets
// (RFC 3596 §2.2).
func (rr *ResourceRecord) AsAAAA() (net.IP, error) {
	if rr.Type != llmnr_type.TypeAAAA {
		return nil, fmt.Errorf("record type is %s, not AAAA", rr.Type.String())
	}
	if len(rr.RData) != net.IPv6len {
		return nil, fmt.Errorf("AAAA rdata length is %d, want %d", len(rr.RData), net.IPv6len)
	}
	ip := make(net.IP, net.IPv6len)
	copy(ip, rr.RData)
	return ip, nil
}

// AsPTR returns the domain name carried by a PTR record (RFC 1035 §3.3.12).
func (rr *ResourceRecord) AsPTR() (string, error) {
	if rr.Type != llmnr_type.TypePTR {
		return "", fmt.Errorf("record type is %s, not PTR", rr.Type.String())
	}
	return rr.decodeName(0)
}

// AsCNAME returns the canonical name carried by a CNAME record (RFC 1035 §3.3.1).
func (rr *ResourceRecord) AsCNAME() (string, error) {
	if rr.Type != llmnr_type.TypeCNAME {
		return "", fmt.Errorf("record type is %s, not CNAME", rr.Type.String())
	}
	return rr.decodeName(0)
}

// AsNS returns the name server carried by an NS record (RFC 1035 §3.3.11).
func (rr *ResourceRecord) AsNS() (string, error) {
	if rr.Type != llmnr_type.TypeNS {
		return "", fmt.Errorf("record type is %s, not NS", rr.Type.String())
	}
	return rr.decodeName(0)
}

// AsTXT returns the character-strings carried by a TXT record. TXT RDATA is one
// or more <character-string>s, each a single length octet followed by that many
// bytes (RFC 1035 §3.3.14 and §3.3).
func (rr *ResourceRecord) AsTXT() ([]string, error) {
	if rr.Type != llmnr_type.TypeTXT {
		return nil, fmt.Errorf("record type is %s, not TXT", rr.Type.String())
	}

	strs := []string{}
	offset := 0
	for offset < len(rr.RData) {
		length := int(rr.RData[offset])
		offset++
		if offset+length > len(rr.RData) {
			return nil, fmt.Errorf("truncated txt character-string")
		}
		strs = append(strs, string(rr.RData[offset:offset+length]))
		offset += length
	}
	return strs, nil
}

// AsSRV returns the priority, weight, port and target of an SRV record. SRV
// RDATA is three big-endian uint16s (priority, weight, port) followed by the
// target domain name (RFC 2782). The target may in principle contain a
// compression pointer, which is resolved against the message origin.
func (rr *ResourceRecord) AsSRV() (priority uint16, weight uint16, port uint16, target string, err error) {
	if rr.Type != llmnr_type.TypeSRV {
		return 0, 0, 0, "", fmt.Errorf("record type is %s, not SRV", rr.Type.String())
	}
	if len(rr.RData) < 6 {
		return 0, 0, 0, "", fmt.Errorf("srv rdata length is %d, want at least 6", len(rr.RData))
	}

	priority = binary.BigEndian.Uint16(rr.RData[0:2])
	weight = binary.BigEndian.Uint16(rr.RData[2:4])
	port = binary.BigEndian.Uint16(rr.RData[4:6])

	target, err = rr.decodeName(6)
	if err != nil {
		return 0, 0, 0, "", err
	}
	return priority, weight, port, target, nil
}

// NameToRData encodes a single domain name as RDATA for a name-bearing record
// (PTR/CNAME/NS). The name is emitted uncompressed; compression-on-write is a
// separate concern and Marshal recomputes RDLENGTH from the resulting bytes.
func NameToRData(name string) ([]byte, error) {
	return domain_name.EncodeDomainName(name)
}

// TXTToRData encodes one or more character-strings as TXT RDATA (RFC 1035
// §3.3.14). Each string is emitted as a single length octet followed by its
// bytes; a string longer than 255 bytes cannot be represented and is rejected.
func TXTToRData(strs []string) ([]byte, error) {
	data := []byte{}
	for _, s := range strs {
		if len(s) > 255 {
			return nil, fmt.Errorf("txt character-string too long: %d bytes, max 255", len(s))
		}
		data = append(data, byte(len(s)))
		data = append(data, []byte(s)...)
	}
	return data, nil
}

// SRVToRData encodes an SRV record's RDATA (RFC 2782): the priority, weight and
// port as big-endian uint16s followed by the uncompressed target name.
func SRVToRData(priority uint16, weight uint16, port uint16, target string) ([]byte, error) {
	data := make([]byte, 6)
	binary.BigEndian.PutUint16(data[0:2], priority)
	binary.BigEndian.PutUint16(data[2:4], weight)
	binary.BigEndian.PutUint16(data[4:6], port)

	nameBuf, err := domain_name.EncodeDomainName(target)
	if err != nil {
		return nil, err
	}
	data = append(data, nameBuf...)
	return data, nil
}
