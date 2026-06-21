package regf

import (
	"encoding/binary"
	"fmt"
)

const (
	hbinSignature  = 0x6E696268 // "hbin" in little-endian
	hbinHeaderSize = 32
)

// HiveBin is the 32-byte header of a hive bin block. Each hive bin contains cells
// and is a multiple of 4096 bytes in size.
//
// Source: https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md#hive-bin
type HiveBin struct {
	// Signature (4 bytes): must be ASCII "hbin" (0x6E696268 little-endian).
	Signature uint32

	// Offset (4 bytes): offset of this hive bin from start of hive bins data.
	Offset uint32

	// Size (4 bytes): total size of this hive bin including header, multiple of 4096.
	Size uint32

	// Reserved (8 bytes): unused.
	Reserved uint64

	// Timestamp (8 bytes): FILETIME (UTC); only meaningful for the first hive bin.
	Timestamp uint64

	// Spare (4 bytes): memory allocation field, no disk meaning.
	Spare uint32
}

// NewHiveBin creates a new empty HiveBin.
func NewHiveBin() *HiveBin {
	return &HiveBin{}
}

// Unmarshal deserializes a HiveBin header from binary data.
//
// Parameters:
//   - data ([]byte): at least 32 bytes of raw hive bin header.
//
// Returns:
//   - The number of bytes consumed (always 32).
//   - An error if the data is too short or the signature is invalid.
func (h *HiveBin) Unmarshal(data []byte) (int, error) {
	if len(data) < hbinHeaderSize {
		return 0, fmt.Errorf("data too short for HiveBin: need %d bytes, got %d", hbinHeaderSize, len(data))
	}

	h.Signature = binary.LittleEndian.Uint32(data[0:4])
	if h.Signature != hbinSignature {
		return 0, fmt.Errorf("invalid HiveBin Signature: 0x%08X (expected 0x%08X)", h.Signature, hbinSignature)
	}

	h.Offset = binary.LittleEndian.Uint32(data[4:8])
	h.Size = binary.LittleEndian.Uint32(data[8:12])
	h.Reserved = binary.LittleEndian.Uint64(data[12:20])
	h.Timestamp = binary.LittleEndian.Uint64(data[20:28])
	h.Spare = binary.LittleEndian.Uint32(data[28:32])

	return hbinHeaderSize, nil
}

// Marshal serializes the HiveBin header to binary data.
//
// Returns:
//   - A byte slice of exactly 32 bytes.
//   - An error if serialization fails.
func (h *HiveBin) Marshal() ([]byte, error) {
	buf := make([]byte, hbinHeaderSize)

	binary.LittleEndian.PutUint32(buf[0:4], h.Signature)
	binary.LittleEndian.PutUint32(buf[4:8], h.Offset)
	binary.LittleEndian.PutUint32(buf[8:12], h.Size)
	binary.LittleEndian.PutUint64(buf[12:20], h.Reserved)
	binary.LittleEndian.PutUint64(buf[20:28], h.Timestamp)
	binary.LittleEndian.PutUint32(buf[28:32], h.Spare)

	return buf, nil
}
