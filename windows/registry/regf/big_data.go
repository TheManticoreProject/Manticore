package regf

import (
	"encoding/binary"
	"fmt"
)

const (
	dbSignature = 0x6264 // "db" in little-endian
	dbMinSize   = 8

	// bigDataThreshold is the maximum value-data size that is stored directly in a single
	// data cell. A value whose data exceeds this is stored in a big-data (db) record whose
	// segments each hold up to this many bytes. Big data exists since hive format 1.4.
	bigDataThreshold = 16344
)

// BigData is a parsed DB (big data) record. When a value's data exceeds bigDataThreshold
// bytes, its KeyValue.DataOffset points to one of these instead of a single data cell; the
// record references a list of data-segment cells that, concatenated, form the value data.
//
// Source: https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md#big-data
type BigData struct {
	// Signature (2 bytes): must be ASCII "db" (0x6264 little-endian).
	Signature uint16

	// NumberOfSegments (2 bytes): count of data segments.
	NumberOfSegments uint16

	// SegmentsListOffset (4 bytes): offset to a cell holding NumberOfSegments 4-byte
	// offsets, each pointing to a data-segment cell.
	SegmentsListOffset uint32
}

// NewBigData creates a new empty BigData.
func NewBigData() *BigData {
	return &BigData{}
}

// Unmarshal deserializes a BigData record from cell data.
//
// Parameters:
//   - data ([]byte): cell data starting with the "db" signature.
//
// Returns:
//   - The number of bytes consumed (always 8).
//   - An error if the data is too short or the signature is invalid.
func (d *BigData) Unmarshal(data []byte) (int, error) {
	if len(data) < dbMinSize {
		return 0, fmt.Errorf("data too short for BigData: need %d bytes, got %d", dbMinSize, len(data))
	}

	d.Signature = binary.LittleEndian.Uint16(data[0:2])
	if d.Signature != dbSignature {
		return 0, fmt.Errorf("invalid BigData Signature: 0x%04X (expected 0x%04X)", d.Signature, dbSignature)
	}

	d.NumberOfSegments = binary.LittleEndian.Uint16(data[2:4])
	d.SegmentsListOffset = binary.LittleEndian.Uint32(data[4:8])

	return dbMinSize, nil
}

// Marshal serializes the BigData record to binary data.
//
// Returns:
//   - A byte slice of exactly 8 bytes.
//   - An error if serialization fails.
func (d *BigData) Marshal() ([]byte, error) {
	buf := make([]byte, dbMinSize)
	binary.LittleEndian.PutUint16(buf[0:2], d.Signature)
	binary.LittleEndian.PutUint16(buf[2:4], d.NumberOfSegments)
	binary.LittleEndian.PutUint32(buf[4:8], d.SegmentsListOffset)
	return buf, nil
}
