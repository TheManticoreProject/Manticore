// Package regf implements a read-only parser for Windows registry hive files in the
// REGF binary format. It supports offline parsing of SAM, SYSTEM, SECURITY, and other
// hive files for credential extraction and forensic analysis.
//
// References:
//   - https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md
//   - https://github.com/libyal/libregf/blob/main/documentation/Windows%20NT%20Registry%20File%20(REGF)%20format.asciidoc
//   - https://projectzero.google/2024/12/the-windows-registry-adventure-5-regf.html
package regf

import (
	"encoding/binary"
	"fmt"
)

const (
	regfSignature = 0x66676572 // "regf" in little-endian
	baseBlockSize = 4096
)

// BaseBlock is the 4096-byte file header of a REGF hive file.
//
// Source: https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md#base-block
type BaseBlock struct {
	// Signature (4 bytes): must be ASCII "regf" (0x66676572 little-endian).
	Signature uint32

	// PrimarySequenceNumber (4 bytes): incremented at the start of a write operation.
	PrimarySequenceNumber uint32

	// SecondarySequenceNumber (4 bytes): incremented at the end of a write operation.
	SecondarySequenceNumber uint32

	// LastWrittenTimestamp (8 bytes): FILETIME (UTC).
	LastWrittenTimestamp uint64

	// MajorVersion (4 bytes): always 1.
	MajorVersion uint32

	// MinorVersion (4 bytes): 3, 4, 5, or 6.
	MinorVersion uint32

	// FileType (4 bytes): 0 = primary file.
	FileType uint32

	// FileFormat (4 bytes): 1 = direct memory load.
	FileFormat uint32

	// RootCellOffset (4 bytes): offset of the root key node cell, relative from start of
	// hive bins data.
	RootCellOffset uint32

	// HiveBinsDataSize (4 bytes): total size of all hive bins data in bytes.
	HiveBinsDataSize uint32

	// ClusteringFactor (4 bytes): disk sector size / 512.
	ClusteringFactor uint32

	// FileName (64 bytes): UTF-16LE string, partial path of primary file.
	FileName [64]byte

	// Checksum (4 bytes): XOR-32 of first 508 bytes.
	Checksum uint32
}

// NewBaseBlock creates a new empty BaseBlock.
func NewBaseBlock() *BaseBlock {
	return &BaseBlock{}
}

// Unmarshal deserializes a BaseBlock from binary data.
//
// Parameters:
//   - data ([]byte): at least 4096 bytes of raw hive file header.
//
// Returns:
//   - The number of bytes consumed.
//   - An error if the data is too short or the signature is invalid.
func (b *BaseBlock) Unmarshal(data []byte) (int, error) {
	if len(data) < baseBlockSize {
		return 0, fmt.Errorf("data too short for BaseBlock: need %d bytes, got %d", baseBlockSize, len(data))
	}

	b.Signature = binary.LittleEndian.Uint32(data[0:4])
	if b.Signature != regfSignature {
		return 0, fmt.Errorf("invalid BaseBlock Signature: 0x%08X (expected 0x%08X)", b.Signature, regfSignature)
	}

	b.PrimarySequenceNumber = binary.LittleEndian.Uint32(data[4:8])
	b.SecondarySequenceNumber = binary.LittleEndian.Uint32(data[8:12])
	b.LastWrittenTimestamp = binary.LittleEndian.Uint64(data[12:20])
	b.MajorVersion = binary.LittleEndian.Uint32(data[20:24])
	b.MinorVersion = binary.LittleEndian.Uint32(data[24:28])
	b.FileType = binary.LittleEndian.Uint32(data[28:32])
	b.FileFormat = binary.LittleEndian.Uint32(data[32:36])
	b.RootCellOffset = binary.LittleEndian.Uint32(data[36:40])
	b.HiveBinsDataSize = binary.LittleEndian.Uint32(data[40:44])
	b.ClusteringFactor = binary.LittleEndian.Uint32(data[44:48])
	copy(b.FileName[:], data[48:112])
	b.Checksum = binary.LittleEndian.Uint32(data[508:512])

	return baseBlockSize, nil
}

// Marshal serializes the BaseBlock to binary data.
//
// Returns:
//   - A byte slice of exactly 4096 bytes.
//   - An error if serialization fails.
func (b *BaseBlock) Marshal() ([]byte, error) {
	buf := make([]byte, baseBlockSize)

	binary.LittleEndian.PutUint32(buf[0:4], b.Signature)
	binary.LittleEndian.PutUint32(buf[4:8], b.PrimarySequenceNumber)
	binary.LittleEndian.PutUint32(buf[8:12], b.SecondarySequenceNumber)
	binary.LittleEndian.PutUint64(buf[12:20], b.LastWrittenTimestamp)
	binary.LittleEndian.PutUint32(buf[20:24], b.MajorVersion)
	binary.LittleEndian.PutUint32(buf[24:28], b.MinorVersion)
	binary.LittleEndian.PutUint32(buf[28:32], b.FileType)
	binary.LittleEndian.PutUint32(buf[32:36], b.FileFormat)
	binary.LittleEndian.PutUint32(buf[36:40], b.RootCellOffset)
	binary.LittleEndian.PutUint32(buf[40:44], b.HiveBinsDataSize)
	binary.LittleEndian.PutUint32(buf[44:48], b.ClusteringFactor)
	copy(buf[48:112], b.FileName[:])
	binary.LittleEndian.PutUint32(buf[508:512], b.Checksum)

	return buf, nil
}
