package ntlm

import (
	"encoding/binary"
	"errors"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/avpair"
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
