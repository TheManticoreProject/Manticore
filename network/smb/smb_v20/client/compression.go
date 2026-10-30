package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
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

// patternV1PayloadSize is the size of SMB2_COMPRESSION_PATTERN_PAYLOAD_V1
// (MS-SMB2 2.2.42.2.2).
const patternV1PayloadSize = 8

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
	// SMB2_COMPRESSION_PATTERN_PAYLOAD_V1 (MS-SMB2 2.2.42.2.2): Pattern(1) +
	// Reserved1(1) + Reserved2(2) + Repetitions(4). The total is always 8 bytes.
	if uint64(len(data)) > 0xFFFFFFFF {
		return nil
	}
	out := make([]byte, patternV1PayloadSize)
	out[0] = pattern
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(data)))
	return out
}

// DecompressPatternV1 decompresses an SMB2_COMPRESSION_PATTERN_PAYLOAD_V1
// (MS-SMB2 2.2.42.2.2). maxSize is the OriginalCompressedSegmentSize of the
// enclosing SMB2_COMPRESSION_TRANSFORM_HEADER_CHAINED; per MS-SMB2 3.1.5.3
// a Repetitions value greater than it is rejected.
func DecompressPatternV1(data []byte, maxSize uint32) ([]byte, error) {
	if len(data) < patternV1PayloadSize {
		return nil, fmt.Errorf("Pattern_V1 data too short: %d bytes, need %d", len(data), patternV1PayloadSize)
	}
	if data[1] != 0 {
		return nil, fmt.Errorf("Pattern_V1 reserved byte is 0x%02x, want 0x00", data[1])
	}
	pattern := data[0]
	count := binary.LittleEndian.Uint32(data[4:8])
	if count > maxSize || uint64(count) > uint64(math.MaxInt) {
		return nil, fmt.Errorf("Pattern_V1 repetitions %d exceed original segment size %d", count, maxSize)
	}
	return bytes.Repeat([]byte{pattern}, int(count)), nil
}

// Chained compression payload header (MS-SMB2 2.2.42.2.1), without the
// optional OriginalPayloadSize field.
const chainedPayloadHeaderSize = 8

// chainedOriginalPayloadSizeLen is the size of the optional OriginalPayloadSize
// field of SMB2_COMPRESSION_CHAINED_PAYLOAD_HEADER.
const chainedOriginalPayloadSizeLen = 4

// CompressionChainedPayloadHeader is a single segment header in a chained
// compression transform (MS-SMB2 2.2.42.2.1). When the message uses
// SMB2_COMPRESSION_TRANSFORM_HEADER_CHAINED, the payload is a sequence of these
// headers, each followed by its compressed segment.
//
// Wire layout: CompressionAlgorithm(2) + Flags(2) + Length(4) +
// OriginalPayloadSize(4, present only for LZNT1, LZ77, LZ77+Huffman and LZ4).
// Length counts the compressed payload including OriginalPayloadSize when
// present.
type CompressionChainedPayloadHeader struct {
	CompressionAlgorithm uint16
	Flags                uint16
	Length               uint32
	OriginalPayloadSize  uint32
}

// chainedHasOriginalPayloadSize reports whether a chained payload header for
// alg carries the OriginalPayloadSize field (MS-SMB2 2.2.42.2.1).
func chainedHasOriginalPayloadSize(alg uint16) bool {
	switch alg {
	case 0x0001, // SMB2_COMPRESSION_LZNT1
		0x0002, // SMB2_COMPRESSION_LZ77
		0x0003, // SMB2_COMPRESSION_LZ77_HUFFMAN
		0x0005: // SMB2_COMPRESSION_LZ4
		return true
	}
	return false
}

// MarshalCompressionChainedPayloadHeader encodes a chained payload header,
// including OriginalPayloadSize when the algorithm requires it. The caller
// appends the compressed segment data after the returned bytes.
func MarshalCompressionChainedPayloadHeader(h *CompressionChainedPayloadHeader) []byte {
	size := chainedPayloadHeaderSize
	if chainedHasOriginalPayloadSize(h.CompressionAlgorithm) {
		size += chainedOriginalPayloadSizeLen
	}
	buf := make([]byte, size)
	binary.LittleEndian.PutUint16(buf[0:2], h.CompressionAlgorithm)
	binary.LittleEndian.PutUint16(buf[2:4], h.Flags)
	binary.LittleEndian.PutUint32(buf[4:8], h.Length)
	if size > chainedPayloadHeaderSize {
		binary.LittleEndian.PutUint32(buf[8:12], h.OriginalPayloadSize)
	}
	return buf
}

// ParseCompressionChainedPayloadHeader decodes a chained payload header,
// including OriginalPayloadSize when the algorithm carries it.
func ParseCompressionChainedPayloadHeader(data []byte) (*CompressionChainedPayloadHeader, error) {
	if len(data) < chainedPayloadHeaderSize {
		return nil, fmt.Errorf("chained payload header too short: %d bytes, need %d", len(data), chainedPayloadHeaderSize)
	}
	h := &CompressionChainedPayloadHeader{
		CompressionAlgorithm: binary.LittleEndian.Uint16(data[0:2]),
		Flags:                binary.LittleEndian.Uint16(data[2:4]),
		Length:               binary.LittleEndian.Uint32(data[4:8]),
	}
	if chainedHasOriginalPayloadSize(h.CompressionAlgorithm) {
		if len(data) < chainedPayloadHeaderSize+chainedOriginalPayloadSizeLen {
			return nil, fmt.Errorf("chained payload header too short for OriginalPayloadSize: %d bytes, need %d", len(data), chainedPayloadHeaderSize+chainedOriginalPayloadSizeLen)
		}
		if h.Length < chainedOriginalPayloadSizeLen {
			return nil, fmt.Errorf("chained payload Length %d smaller than OriginalPayloadSize field", h.Length)
		}
		h.OriginalPayloadSize = binary.LittleEndian.Uint32(data[8:12])
	}
	return h, nil
}

// CompressionChainedSegment is one segment of a chained compressed payload:
// the per-segment header and the compressed data that follows it.
type CompressionChainedSegment struct {
	Header         CompressionChainedPayloadHeader
	CompressedData []byte
}

// WalkCompressionChain parses the chained payload of a chained compression
// transform. data starts immediately after the 8-byte
// SMB2_COMPRESSION_TRANSFORM_HEADER_CHAINED (ProtocolId +
// OriginalCompressedSegmentSize), i.e. at the first
// SMB2_COMPRESSION_CHAINED_PAYLOAD_HEADER. Each segment spans Length bytes after
// its 8-byte header; for algorithms carrying OriginalPayloadSize, those 4 bytes
// are part of Length and are excluded from CompressedData.
func WalkCompressionChain(data []byte) ([]CompressionChainedSegment, error) {
	var segments []CompressionChainedSegment
	offset := 0
	for offset < len(data) {
		if len(data)-offset < chainedPayloadHeaderSize {
			return nil, fmt.Errorf("chained segment at offset %d too short for header", offset)
		}
		hdr, err := ParseCompressionChainedPayloadHeader(data[offset:])
		if err != nil {
			return nil, err
		}
		start := offset + chainedPayloadHeaderSize
		if uint64(hdr.Length) > uint64(len(data)-start) {
			return nil, fmt.Errorf("chained segment at offset %d: length %d exceeds buffer", offset, hdr.Length)
		}
		end := start + int(hdr.Length)
		if chainedHasOriginalPayloadSize(hdr.CompressionAlgorithm) {
			start += chainedOriginalPayloadSizeLen
		}
		segments = append(segments, CompressionChainedSegment{
			Header:         *hdr,
			CompressedData: data[start:end],
		})
		offset = end
	}
	return segments, nil
}
