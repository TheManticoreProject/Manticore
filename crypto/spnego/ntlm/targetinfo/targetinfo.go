package targetinfo

import (
	"encoding/binary"
	"errors"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// ParseTargetInfo parses the target info from a challenge message
func ParseTargetInfo(targetInfo []byte) (map[avpair.AvId][]byte, error) {
	result := make(map[avpair.AvId][]byte)

	offset := 0
	for offset < len(targetInfo) {
		// Need at least 4 bytes for the AV_PAIR header
		if offset+4 > len(targetInfo) {
			return nil, errors.New("target info truncated")
		}

		avId := avpair.AvId(binary.LittleEndian.Uint16(targetInfo[offset : offset+2]))
		offset += 2

		avLen := binary.LittleEndian.Uint16(targetInfo[offset : offset+2])
		offset += 2

		// Check if we have enough bytes for the value
		if offset+int(avLen) > len(targetInfo) {
			return nil, errors.New("target info value truncated")
		}

		// Extract the value
		if avId != avpair.MsvAvEOL {
			result[avId] = targetInfo[offset : offset+int(avLen)]
		}

		offset += int(avLen)

		// If we reached the end of list marker, we're done
		if avId == avpair.MsvAvEOL {
			break
		}
	}

	return result, nil
}

// HasTimestamp reports whether MsvAvTimestamp is present in the TargetInfo.
func HasTimestamp(targetInfo []byte) bool {
	pairs, err := ParseTargetInfo(targetInfo)
	if err != nil {
		return false
	}
	_, ok := pairs[avpair.MsvAvTimestamp]
	return ok
}

// GetTimestamp returns the raw 8-byte Windows FILETIME from TargetInfo, or nil if absent.
func GetTimestamp(targetInfo []byte) []byte {
	pairs, err := ParseTargetInfo(targetInfo)
	if err != nil {
		return nil
	}
	ts, ok := pairs[avpair.MsvAvTimestamp]
	if !ok {
		return nil
	}
	return ts
}

// BuildBlobTargetInfo constructs the modified TargetInfo to embed in the NTLMv2 blob.
//
// It copies all AVPairs from the challenge TargetInfo and, when a DNS computer name
// (MsvAvDnsComputerName) is present, inserts an MsvAvTargetName AVPair set to the SMB
// service principal name "cifs/<DnsComputerName>" before the EOL marker. This mirrors
// what the Windows client sends; modern Windows servers require the SPN in the
// AUTHENTICATE's NTLMv2 AVPairs and reject the authentication (STATUS_INVALID_PARAMETER)
// when it is absent.
func BuildBlobTargetInfo(targetInfo []byte) []byte {
	// Build the SMB SPN "cifs/<DnsComputerName>" from the DNS computer name AVPair,
	// in UTF-16LE (DnsComputerName is already UTF-16LE on the wire).
	var targetName []byte
	if pairs, err := ParseTargetInfo(targetInfo); err == nil {
		if dnsComputer, ok := pairs[avpair.MsvAvDnsComputerName]; ok && len(dnsComputer) > 0 {
			targetName = append(utf16.EncodeUTF16LE("cifs/"), dnsComputer...)
		}
	}

	result := make([]byte, 0, len(targetInfo)+4+len(targetName))
	i := 0
	for i+4 <= len(targetInfo) {
		currentID := avpair.AvId(binary.LittleEndian.Uint16(targetInfo[i : i+2]))
		avLen := binary.LittleEndian.Uint16(targetInfo[i+2 : i+4])

		if currentID == avpair.MsvAvEOL {
			// Insert MsvAvTargetName (SPN) before EOL, then append EOL.
			if len(targetName) > 0 {
				hdr := make([]byte, 4)
				binary.LittleEndian.PutUint16(hdr[0:2], uint16(avpair.MsvAvTargetName))
				binary.LittleEndian.PutUint16(hdr[2:4], uint16(len(targetName)))
				result = append(result, hdr...)
				result = append(result, targetName...)
			}
			result = append(result, targetInfo[i:i+4]...)
			return result
		}

		// avLen is server-controlled; ensure the declared value bytes are present
		// before slicing, so a length running past the buffer cannot panic. A
		// malformed (truncated) tail stops processing rather than crashing.
		if i+4+int(avLen) > len(targetInfo) {
			break
		}
		result = append(result, targetInfo[i:i+4+int(avLen)]...)
		i += 4 + int(avLen)
	}

	return result
}
