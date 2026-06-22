package regf

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	hvleSignature = 0x454C7648 // "HvLE" in little-endian (new-format log, hive >= 1.5)
	dirtSignature = 0x54524944 // "DIRT" in little-endian (legacy log, hive < 1.5)
	// logEntriesOffset is where the log body begins in a transaction-log (.LOG1/.LOG2)
	// file: the log's REGF header occupies the first 512 bytes. Both the HvLE entry stream
	// and the DIRT signature start here.
	logEntriesOffset = 512
	// logEntryHeaderSize is the fixed part of an HvLE log entry, before the dirty-page
	// reference array.
	logEntryHeaderSize = 40
	// dirtPageSize is the granularity of a DIRT log's dirty-page bitmap: one bit per this
	// many bytes of hive-bins data.
	dirtPageSize = 512
	// dirtDataOffset is where the DIRT log's dirty-page data begins (after the header,
	// signature, and bitmap region).
	dirtDataOffset = 1024
)

// DirtyPageReference locates one dirty page within a log entry: Offset is relative to the
// start of the primary hive's hive-bins data (absolute file position = 4096 + Offset) and
// Size is the page length in bytes.
//
// Source: https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md#dirty-pages-references
type DirtyPageReference struct {
	Offset uint32
	Size   uint32
}

// LogEntry is a parsed HvLE transaction-log entry: a single logged transaction carrying a
// set of dirty pages to be written back into the primary hive during recovery.
//
// Source: https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md#log-entry
type LogEntry struct {
	// Signature (4 bytes): must be ASCII "HvLE" (0x454C7648 little-endian).
	Signature uint32

	// LogSize (4 bytes): total size of this log entry including header and page data,
	// aligned to 512 bytes.
	LogSize uint32

	// Flags (4 bytes): log-entry flags.
	Flags uint32

	// SequenceNumber (4 bytes): the sequence number this entry advances the hive to.
	SequenceNumber uint32

	// HiveBinsDataSize (4 bytes): hive-bins data size after this entry is applied.
	HiveBinsDataSize uint32

	// DirtyPagesCount (4 bytes): number of dirty pages in this entry.
	DirtyPagesCount uint32

	// Hash1 (8 bytes): Marvin64 hash of the dirty-page data (not verified here).
	Hash1 uint64

	// Hash2 (8 bytes): Marvin64 hash of the entry header (not verified here).
	Hash2 uint64

	// DirtyPages are the page references; PageData holds the corresponding page bytes in
	// the same order.
	DirtyPages []DirtyPageReference
	PageData   [][]byte
}

// NewLogEntry creates a new empty LogEntry.
func NewLogEntry() *LogEntry {
	return &LogEntry{}
}

// Unmarshal deserializes a LogEntry from data starting at its "HvLE" signature.
//
// Returns:
//   - The number of bytes consumed (the entry's LogSize, so the caller can advance to the
//     next entry).
//   - An error if the data is too short, the signature is invalid, or a referenced page
//     runs past the buffer.
func (e *LogEntry) Unmarshal(data []byte) (int, error) {
	if len(data) < logEntryHeaderSize {
		return 0, fmt.Errorf("data too short for LogEntry: need %d bytes, got %d", logEntryHeaderSize, len(data))
	}

	e.Signature = binary.LittleEndian.Uint32(data[0:4])
	if e.Signature != hvleSignature {
		return 0, fmt.Errorf("invalid LogEntry Signature: 0x%08X (expected 0x%08X)", e.Signature, hvleSignature)
	}

	e.LogSize = binary.LittleEndian.Uint32(data[4:8])
	e.Flags = binary.LittleEndian.Uint32(data[8:12])
	e.SequenceNumber = binary.LittleEndian.Uint32(data[12:16])
	e.HiveBinsDataSize = binary.LittleEndian.Uint32(data[16:20])
	e.DirtyPagesCount = binary.LittleEndian.Uint32(data[20:24])
	e.Hash1 = binary.LittleEndian.Uint64(data[24:32])
	e.Hash2 = binary.LittleEndian.Uint64(data[32:40])

	pos := logEntryHeaderSize
	refsEnd := pos + int(e.DirtyPagesCount)*8
	if refsEnd > len(data) {
		return 0, fmt.Errorf("LogEntry dirty-page references run past buffer (%d > %d)", refsEnd, len(data))
	}
	e.DirtyPages = make([]DirtyPageReference, e.DirtyPagesCount)
	for i := range e.DirtyPages {
		e.DirtyPages[i].Offset = binary.LittleEndian.Uint32(data[pos : pos+4])
		e.DirtyPages[i].Size = binary.LittleEndian.Uint32(data[pos+4 : pos+8])
		pos += 8
	}

	e.PageData = make([][]byte, e.DirtyPagesCount)
	for i, ref := range e.DirtyPages {
		end := pos + int(ref.Size)
		if end > len(data) {
			return 0, fmt.Errorf("LogEntry dirty page %d (size %d) runs past buffer", i, ref.Size)
		}
		e.PageData[i] = append([]byte(nil), data[pos:end]...)
		pos += int(ref.Size)
	}

	consumed := int(e.LogSize)
	if consumed < pos {
		// LogSize must at least cover the header, references, and page data.
		return 0, fmt.Errorf("LogEntry LogSize %d smaller than its content %d", e.LogSize, pos)
	}
	return consumed, nil
}

// Marshal serializes the LogEntry: header, dirty-page references, page data, then zero
// padding up to LogSize. DirtyPagesCount and LogSize must be set consistently with
// DirtyPages/PageData.
func (e *LogEntry) Marshal() ([]byte, error) {
	body := logEntryHeaderSize + len(e.DirtyPages)*8
	for _, p := range e.PageData {
		body += len(p)
	}
	size := int(e.LogSize)
	if size < body {
		size = body
	}
	buf := make([]byte, size)

	binary.LittleEndian.PutUint32(buf[0:4], e.Signature)
	binary.LittleEndian.PutUint32(buf[4:8], e.LogSize)
	binary.LittleEndian.PutUint32(buf[8:12], e.Flags)
	binary.LittleEndian.PutUint32(buf[12:16], e.SequenceNumber)
	binary.LittleEndian.PutUint32(buf[16:20], e.HiveBinsDataSize)
	binary.LittleEndian.PutUint32(buf[20:24], e.DirtyPagesCount)
	binary.LittleEndian.PutUint64(buf[24:32], e.Hash1)
	binary.LittleEndian.PutUint64(buf[32:40], e.Hash2)

	pos := logEntryHeaderSize
	for _, ref := range e.DirtyPages {
		binary.LittleEndian.PutUint32(buf[pos:pos+4], ref.Offset)
		binary.LittleEndian.PutUint32(buf[pos+4:pos+8], ref.Size)
		pos += 8
	}
	for _, p := range e.PageData {
		copy(buf[pos:], p)
		pos += len(p)
	}
	return buf, nil
}

// TransactionLog is a parsed transaction-log file: the sequence of HvLE entries that
// follow the 512-byte log header.
type TransactionLog struct {
	Entries []*LogEntry
}

// Unmarshal parses the HvLE entries from a transaction-log file's bytes. Parsing stops at
// the first non-HvLE object (a malformed or empty trailing entry), which is normal: a log
// file commonly contains stale space after its last valid entry.
func (t *TransactionLog) Unmarshal(data []byte) error {
	pos := logEntriesOffset
	for pos+4 <= len(data) {
		if binary.LittleEndian.Uint32(data[pos:pos+4]) != hvleSignature {
			break
		}
		e := NewLogEntry()
		n, err := e.Unmarshal(data[pos:])
		if err != nil {
			break // trailing garbage / partial entry: stop at the last good entry
		}
		t.Entries = append(t.Entries, e)
		if n <= 0 {
			break
		}
		pos += n
	}
	return nil
}

// ReplayTransactionLog applies a transaction log to a primary hive image and returns the
// recovered image plus the number of entries applied. It applies entries whose sequence
// numbers run contiguously from the hive's secondary sequence number, stopping at the
// first gap (the standard recovery rule); each entry's dirty pages are written at
// 4096+Offset, the image is grown if a page extends it, and the base block's sequence
// numbers and hive-bins-data size are updated to reflect the applied state.
func ReplayTransactionLog(hiveData, logData []byte) ([]byte, int, error) {
	var bb BaseBlock
	if _, err := bb.Unmarshal(hiveData); err != nil {
		return nil, 0, fmt.Errorf("regf: replay: %w", err)
	}
	image := append([]byte(nil), hiveData...)

	// Dispatch on the log magic at offset 512: legacy DIRT vs modern HvLE.
	if len(logData) >= logEntriesOffset+4 &&
		binary.LittleEndian.Uint32(logData[logEntriesOffset:logEntriesOffset+4]) == dirtSignature {
		return replayDirtLog(image, logData, bb.HiveBinsDataSize)
	}

	var tl TransactionLog
	if err := tl.Unmarshal(logData); err != nil {
		return nil, 0, fmt.Errorf("regf: replay: %w", err)
	}

	expected := bb.SecondarySequenceNumber
	applied := 0
	for _, e := range tl.Entries {
		if e.SequenceNumber != expected {
			break // gap in the sequence: cannot apply further
		}
		for i, ref := range e.DirtyPages {
			abs := baseBlockSize + int(ref.Offset)
			end := abs + len(e.PageData[i])
			if end > len(image) {
				image = append(image, make([]byte, end-len(image))...)
			}
			copy(image[abs:end], e.PageData[i])
		}
		expected++
		// Mirror the applied state into the base block: sequence numbers (offsets 4 and 8)
		// and hive-bins-data size (offset 40).
		binary.LittleEndian.PutUint32(image[4:8], expected)
		binary.LittleEndian.PutUint32(image[8:12], expected)
		binary.LittleEndian.PutUint32(image[40:44], e.HiveBinsDataSize)
		applied++
	}
	return image, applied, nil
}

// replayDirtLog applies a legacy "DIRT" transaction log to the primary image. The log
// carries a dirty-page bitmap (hbinsDataSize/dirtPageSize bits, one per 512-byte block of
// hive-bins data) at offset 516; each set bit's 512-byte block is packed, in bitmap order,
// starting at dirtDataOffset and is written to 4096 + bitIndex*512 in the primary. Unlike
// HvLE, the base block is left untouched (matching the reference implementation).
func replayDirtLog(image, logData []byte, hbinsDataSize uint32) ([]byte, int, error) {
	bitmapLen := int(hbinsDataSize) / dirtPageSize / 8
	if bitmapLen == 0 {
		return image, 0, nil
	}
	bitmapStart := logEntriesOffset + 4 // after the "DIRT" signature
	if bitmapStart+bitmapLen > len(logData) {
		return nil, 0, fmt.Errorf("regf: DIRT log truncated: need %d-byte dirty bitmap", bitmapLen)
	}
	bitmap := logData[bitmapStart : bitmapStart+bitmapLen]

	applied := 0
	setIndex := 0
	for bit := 0; bit < bitmapLen*8; bit++ {
		if bitmap[bit/8]>>(uint(bit)%8)&1 == 0 {
			continue
		}
		src := dirtDataOffset + setIndex*dirtPageSize
		dst := baseBlockSize + bit*dirtPageSize
		if src+dirtPageSize > len(logData) {
			return nil, 0, fmt.Errorf("regf: DIRT log truncated at dirty block %d", setIndex)
		}
		if dst+dirtPageSize > len(image) {
			image = append(image, make([]byte, dst+dirtPageSize-len(image))...)
		}
		copy(image[dst:dst+dirtPageSize], logData[src:src+dirtPageSize])
		setIndex++
		applied++
	}
	return image, applied, nil
}

// OpenWithLogs opens a primary hive and replays the given transaction logs (in order,
// typically .LOG1 then .LOG2) before parsing, returning a Hive that reflects the recovered
// state. With no log paths it is equivalent to Open.
func OpenWithLogs(hivePath string, logPaths ...string) (*Hive, error) {
	data, err := os.ReadFile(hivePath)
	if err != nil {
		return nil, fmt.Errorf("regf: open %s: %w", hivePath, err)
	}
	for _, lp := range logPaths {
		logData, err := os.ReadFile(lp)
		if err != nil {
			return nil, fmt.Errorf("regf: open log %s: %w", lp, err)
		}
		recovered, _, err := ReplayTransactionLog(data, logData)
		if err != nil {
			return nil, fmt.Errorf("regf: replay %s: %w", lp, err)
		}
		data = recovered
	}
	return OpenBytes(data)
}
