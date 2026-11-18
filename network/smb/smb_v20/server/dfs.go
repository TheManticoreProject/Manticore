package server

import (
	"encoding/binary"
	"strings"
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

const dfsReferralSiteNameFlag uint16 = 0x0001

// dfsReferralStatus processes a referral request for this server's empty DFS
// namespace list. A well-formed stand-alone root or link is therefore absent;
// domain and DC discovery requests are invalid because this server is not a DC.
func (c *Connection) dfsReferralStatus(input []byte, extended bool) nt_status.NT_STATUS {
	path, ok := parseDFSReferralRequest(input, extended)
	if !ok {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}
	if path == "" {
		// An empty path is a domain referral request.
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	hasLeadingSlash := strings.HasPrefix(path, `\`)
	if hasLeadingSlash {
		path = strings.TrimPrefix(path, `\`)
		if path == "" || strings.HasPrefix(path, `\`) {
			return nt_status.NT_STATUS_INVALID_PARAMETER
		}
	}
	components := strings.Split(path, `\`)
	for _, component := range components {
		if component == "" {
			return nt_status.NT_STATUS_INVALID_PARAMETER
		}
	}
	if len(components) == 1 {
		// One component, with or without the leading slash, is a DC referral.
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}
	if !hasLeadingSlash {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	// A non-DC has no SYSVOL or NETLOGON referrals.
	if strings.EqualFold(components[1], "SYSVOL") || strings.EqualFold(components[1], "NETLOGON") {
		return nt_status.NT_STATUS_NOT_FOUND
	}

	// An absent domain-based namespace is unavailable; an absent stand-alone
	// namespace is not found (MS-DFSC 3.2.5.5).
	config := c.server.config
	if strings.EqualFold(components[0], config.DomainName) || strings.EqualFold(components[0], config.DNSDomainName) {
		return nt_status.NT_STATUS_DFS_UNAVAILABLE
	}
	return nt_status.NT_STATUS_NOT_FOUND
}

func parseDFSReferralRequest(input []byte, extended bool) (string, bool) {
	if len(input) < 2 || binary.LittleEndian.Uint16(input[:2]) == 0 {
		return "", false
	}
	if !extended {
		return decodeDFSString(input[2:])
	}
	if len(input) < 10 {
		return "", false
	}
	flags := binary.LittleEndian.Uint16(input[2:4])
	if flags & ^dfsReferralSiteNameFlag != 0 {
		return "", false
	}
	dataLength := int(binary.LittleEndian.Uint32(input[4:8]))
	if dataLength != len(input)-8 {
		return "", false
	}

	data := input[8:]
	pathLength := int(binary.LittleEndian.Uint16(data[:2]))
	if pathLength > len(data)-2 {
		return "", false
	}
	path, ok := decodeDFSString(data[2 : 2+pathLength])
	if !ok {
		return "", false
	}
	data = data[2+pathLength:]

	if flags&dfsReferralSiteNameFlag == 0 {
		return path, len(data) == 0
	}
	if len(data) < 2 {
		return "", false
	}
	siteLength := int(binary.LittleEndian.Uint16(data[:2]))
	if siteLength != len(data)-2 {
		return "", false
	}
	_, ok = decodeDFSString(data[2:])
	return path, ok
}

func decodeDFSString(data []byte) (string, bool) {
	if len(data) < 2 || len(data)%2 != 0 || binary.LittleEndian.Uint16(data[len(data)-2:]) != 0 {
		return "", false
	}
	units := make([]uint16, 0, len(data)/2-1)
	for offset := 0; offset < len(data)-2; offset += 2 {
		unit := binary.LittleEndian.Uint16(data[offset:])
		if unit == 0 {
			return "", false
		}
		units = append(units, unit)
	}
	return string(utf16.Decode(units)), true
}
