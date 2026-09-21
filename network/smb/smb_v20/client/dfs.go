package client

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
)

// DFS referral header flags (MS-DFSC 2.2.4).
const (
	DFS_REFERRAL_HEADER_FLAG_REFERRAL_SERVERS uint32 = 0x00000001
	DFS_REFERRAL_HEADER_FLAG_STORAGE_SERVERS  uint32 = 0x00000002
	DFS_REFERRAL_HEADER_FLAG_TARGET_FAILBACK  uint32 = 0x00000004
)

// DFS referral server types (MS-DFSC 2.2.5.3).
const (
	DFS_SERVER_ROOT uint16 = 0x0001
	DFS_SERVER_LINK uint16 = 0x0000
)

// DfsReferralResponse is the parsed RESP_GET_DFS_REFERRAL (MS-DFSC 2.2.4).
type DfsReferralResponse struct {
	PathConsumed  uint16
	HeaderFlags   uint32
	ReferralEntries []DfsReferralEntry
}

// DfsReferralEntry is a single DFS_REFERRAL_V3 or V4 entry.
type DfsReferralEntry struct {
	Version        uint16
	ServerType     uint16
	Flags          uint16
	TimeToLive     uint32
	DFSPath        string
	DFSAlternatePath string
	NetworkAddress string
}

// IsRootTarget reports whether this referral points to a DFS root target.
func (e *DfsReferralEntry) IsRootTarget() bool { return e.ServerType == DFS_SERVER_ROOT }

// marshalDfsReferralRequest encodes a REQ_GET_DFS_REFERRAL (MS-DFSC 2.2.2).
// MaxReferralLevel is typically 4 (highest); requestPath is the UNC path.
func marshalDfsReferralRequest(maxLevel uint16, requestPath string) []byte {
	pathUTF16 := utf16.Encode([]rune(requestPath))
	// MaxReferralLevel(2) + path in UTF-16LE + null terminator(2)
	buf := make([]byte, 2+len(pathUTF16)*2+2)
	binary.LittleEndian.PutUint16(buf[0:2], maxLevel)
	for i, r := range pathUTF16 {
		binary.LittleEndian.PutUint16(buf[2+i*2:], r)
	}
	return buf
}

// parseDfsReferralResponse decodes a RESP_GET_DFS_REFERRAL (MS-DFSC 2.2.4).
func parseDfsReferralResponse(data []byte) (*DfsReferralResponse, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("DFS referral response too short: %d bytes, need at least 8", len(data))
	}

	resp := &DfsReferralResponse{
		PathConsumed: binary.LittleEndian.Uint16(data[0:2]),
		HeaderFlags:  binary.LittleEndian.Uint32(data[4:8]),
	}
	numReferrals := int(binary.LittleEndian.Uint16(data[2:4]))

	offset := 8
	for i := 0; i < numReferrals; i++ {
		if offset+8 > len(data) {
			return nil, fmt.Errorf("DFS referral entry %d truncated at offset %d", i, offset)
		}
		version := binary.LittleEndian.Uint16(data[offset : offset+2])
		entrySize := int(binary.LittleEndian.Uint16(data[offset+2 : offset+4]))
		if entrySize < 8 || offset+entrySize > len(data) {
			return nil, fmt.Errorf("DFS referral entry %d size %d out of bounds at offset %d", i, entrySize, offset)
		}

		entry := DfsReferralEntry{Version: version}
		entryData := data[offset:]

		switch version {
		case 1:
			if entrySize < 8 {
				return nil, fmt.Errorf("DFS V1 referral too short: %d", entrySize)
			}
			entry.ServerType = binary.LittleEndian.Uint16(entryData[4:6])
			// V1 path is at offset 8, null-terminated UTF-16LE
			if offset+8 < len(data) {
				entry.NetworkAddress = readUTF16NullTerm(entryData[8:])
			}

		case 2:
			if entrySize < 22 {
				return nil, fmt.Errorf("DFS V2 referral too short: %d", entrySize)
			}
			entry.ServerType = binary.LittleEndian.Uint16(entryData[4:6])
			entry.TimeToLive = binary.LittleEndian.Uint32(entryData[8:12])
			dfsPathOff := int(binary.LittleEndian.Uint16(entryData[12:14]))
			dfsAltOff := int(binary.LittleEndian.Uint16(entryData[14:16]))
			netAddrOff := int(binary.LittleEndian.Uint16(entryData[16:18]))
			if dfsPathOff > 0 && dfsPathOff < len(entryData) {
				entry.DFSPath = readUTF16NullTerm(entryData[dfsPathOff:])
			}
			if dfsAltOff > 0 && dfsAltOff < len(entryData) {
				entry.DFSAlternatePath = readUTF16NullTerm(entryData[dfsAltOff:])
			}
			if netAddrOff > 0 && netAddrOff < len(entryData) {
				entry.NetworkAddress = readUTF16NullTerm(entryData[netAddrOff:])
			}

		case 3, 4:
			if entrySize < 34 {
				return nil, fmt.Errorf("DFS V%d referral too short: %d", version, entrySize)
			}
			entry.ServerType = binary.LittleEndian.Uint16(entryData[4:6])
			entry.Flags = binary.LittleEndian.Uint16(entryData[6:8])
			entry.TimeToLive = binary.LittleEndian.Uint32(entryData[8:12])
			dfsPathOff := int(binary.LittleEndian.Uint16(entryData[12:14]))
			dfsAltOff := int(binary.LittleEndian.Uint16(entryData[14:16]))
			netAddrOff := int(binary.LittleEndian.Uint16(entryData[16:18]))
			if dfsPathOff > 0 && dfsPathOff < len(entryData) {
				entry.DFSPath = readUTF16NullTerm(entryData[dfsPathOff:])
			}
			if dfsAltOff > 0 && dfsAltOff < len(entryData) {
				entry.DFSAlternatePath = readUTF16NullTerm(entryData[dfsAltOff:])
			}
			if netAddrOff > 0 && netAddrOff < len(entryData) {
				entry.NetworkAddress = readUTF16NullTerm(entryData[netAddrOff:])
			}

		default:
			// Skip unknown versions gracefully.
		}

		resp.ReferralEntries = append(resp.ReferralEntries, entry)
		offset += entrySize
	}
	return resp, nil
}

// readUTF16NullTerm reads a null-terminated UTF-16LE string from data.
func readUTF16NullTerm(data []byte) string {
	var runes []uint16
	for i := 0; i+1 < len(data); i += 2 {
		ch := binary.LittleEndian.Uint16(data[i : i+2])
		if ch == 0 {
			break
		}
		runes = append(runes, ch)
	}
	return string(utf16.Decode(runes))
}

// GetDfsReferral queries the server for DFS referrals for the given path.
// maxLevel is the maximum referral version to request (typically 4).
// The caller must have an IPC$ tree connect established.
func (c *Client) GetDfsReferral(path string, maxLevel uint16) (*DfsReferralResponse, error) {
	fileId := types.SMB2_FILEID{
		Persistent: 0xFFFFFFFFFFFFFFFF,
		Volatile:   0xFFFFFFFFFFFFFFFF,
	}

	input := marshalDfsReferralRequest(maxLevel, path)
	output, err := c.Ioctl(fileId, filesystem.FSCTL_DFS_GET_REFERRALS, input, true, 65536)
	if err != nil {
		return nil, fmt.Errorf("FSCTL_DFS_GET_REFERRALS failed: %w", err)
	}
	return parseDfsReferralResponse(output)
}
