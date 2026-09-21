package client

import (
	"encoding/binary"
	"fmt"
)

// Compression transform header protocol ID (MS-SMB2 2.2.42).
var compressionProtocolId = [4]byte{0xFC, 'S', 'M', 'B'}

// Compression transform flags (MS-SMB2 2.2.42).
const (
	SMB2_COMPRESSION_FLAG_NONE    uint16 = 0x0000
	SMB2_COMPRESSION_FLAG_CHAINED uint16 = 0x0001
)

// compressionTransformHeaderSize is the fixed size of the non-chained
// SMB2_COMPRESSION_TRANSFORM_HEADER.
const compressionTransformHeaderSize = 16

// CompressionTransformHeader is the non-chained SMB2 compression transform
// header (MS-SMB2 2.2.42). When Flags is 0x0000 (non-chained), the header
// precedes a single compressed payload.
type CompressionTransformHeader struct {
	OriginalCompressedSegmentSize uint32
	CompressionAlgorithm          uint16
	Flags                         uint16
	Offset                        uint32
}

// MarshalCompressionTransformHeader encodes a non-chained compression transform
// header. The returned bytes do NOT include the payload — the caller appends the
// uncompressed prefix (Offset bytes) and the compressed segment.
func MarshalCompressionTransformHeader(h *CompressionTransformHeader) []byte {
	buf := make([]byte, compressionTransformHeaderSize)
	copy(buf[0:4], compressionProtocolId[:])
	binary.LittleEndian.PutUint32(buf[4:8], h.OriginalCompressedSegmentSize)
	binary.LittleEndian.PutUint16(buf[8:10], h.CompressionAlgorithm)
	binary.LittleEndian.PutUint16(buf[10:12], h.Flags)
	binary.LittleEndian.PutUint32(buf[12:16], h.Offset)
	return buf
}

// ParseCompressionTransformHeader decodes a non-chained compression transform
// header from data, which must start with the 0xFC 'S' 'M' 'B' protocol ID.
func ParseCompressionTransformHeader(data []byte) (*CompressionTransformHeader, error) {
	if len(data) < compressionTransformHeaderSize {
		return nil, fmt.Errorf("compression transform header too short: %d bytes, need %d", len(data), compressionTransformHeaderSize)
	}
	if data[0] != 0xFC || data[1] != 'S' || data[2] != 'M' || data[3] != 'B' {
		return nil, fmt.Errorf("invalid compression transform protocol ID: %02x %02x %02x %02x", data[0], data[1], data[2], data[3])
	}
	return &CompressionTransformHeader{
		OriginalCompressedSegmentSize: binary.LittleEndian.Uint32(data[4:8]),
		CompressionAlgorithm:          binary.LittleEndian.Uint16(data[8:10]),
		Flags:                         binary.LittleEndian.Uint16(data[10:12]),
		Offset:                        binary.LittleEndian.Uint32(data[12:16]),
	}, nil
}

// IsCompressedMessage checks whether data begins with the SMB2 compression
// transform protocol ID (0xFC 'S' 'M' 'B').
func IsCompressedMessage(data []byte) bool {
	return len(data) >= 4 && data[0] == 0xFC && data[1] == 'S' && data[2] == 'M' && data[3] == 'B'
}

// CompressPatternV1 compresses data using the Pattern_V1 algorithm (MS-SMB2
// 3.1.4.4.1). Pattern_V1 is effective only when the data consists of a single
// repeated byte; otherwise it returns nil to indicate no compression benefit.
func CompressPatternV1(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	pattern := data[0]
	for _, b := range data[1:] {
		if b != pattern {
			return nil
		}
	}
	// Pattern_V1 compressed format: Pattern(1) + Reserved(1) + Repetitions(2)
	// The total is always 4 bytes regardless of input size (up to 65535).
	if len(data) > 0xFFFF {
		return nil
	}
	out := make([]byte, 4)
	out[0] = pattern
	binary.LittleEndian.PutUint16(out[2:4], uint16(len(data)))
	return out
}

// DecompressPatternV1 decompresses a Pattern_V1 compressed payload.
func DecompressPatternV1(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("Pattern_V1 data too short: %d bytes, need 4", len(data))
	}
	pattern := data[0]
	count := int(binary.LittleEndian.Uint16(data[2:4]))
	out := make([]byte, count)
	for i := range out {
		out[i] = pattern
	}
	return out, nil
}
