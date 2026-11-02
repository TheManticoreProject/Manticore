package filesystem

import (
	"encoding/binary"
	"fmt"
)

// FileStreamInformation is FILE_STREAM_INFORMATION (FileInformationClass 22):
// one data stream of a file. A query returns one entry per stream, chained by
// NextEntryOffset and each aligned on an 8-byte boundary.
//
// [MS-FSCC] FileStreamInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/f8762be6-3ab9-411e-a7d6-5cc68f70c78d
type FileStreamInformation struct {
	NextEntryOffset      uint32
	StreamSize           int64
	StreamAllocationSize int64
	// StreamName has the form ":name:$DATA", or "::$DATA" for the unnamed
	// default stream. It is not NUL-terminated on the wire.
	StreamName string
}

// fileStreamInfoFixedSize is the fixed portion preceding StreamName.
const fileStreamInfoFixedSize = 24

// Marshal encodes a single entry (fixed prefix + StreamName). NextEntryOffset
// is written verbatim; MarshalFileStreamInformationList chains entries.
func (fi *FileStreamInformation) Marshal() ([]byte, error) {
	name := encodeUTF16LE(fi.StreamName)
	b := make([]byte, fileStreamInfoFixedSize+len(name))
	binary.LittleEndian.PutUint32(b[0:4], fi.NextEntryOffset)
	binary.LittleEndian.PutUint32(b[4:8], uint32(len(name)))
	binary.LittleEndian.PutUint64(b[8:16], uint64(fi.StreamSize))
	binary.LittleEndian.PutUint64(b[16:24], uint64(fi.StreamAllocationSize))
	copy(b[fileStreamInfoFixedSize:], name)
	return b, nil
}

// Unmarshal decodes a single entry from the start of data.
func (fi *FileStreamInformation) Unmarshal(data []byte) error {
	if len(data) < fileStreamInfoFixedSize {
		return fmt.Errorf("FILE_STREAM_INFORMATION: need %d bytes, got %d", fileStreamInfoFixedSize, len(data))
	}
	*fi = FileStreamInformation{
		NextEntryOffset:      binary.LittleEndian.Uint32(data[0:4]),
		StreamSize:           int64(binary.LittleEndian.Uint64(data[8:16])),
		StreamAllocationSize: int64(binary.LittleEndian.Uint64(data[16:24])),
	}
	nameLen := int(binary.LittleEndian.Uint32(data[4:8]))
	if nameLen > 0 && nameLen <= len(data)-fileStreamInfoFixedSize {
		fi.StreamName = decodeUTF16LE(data[fileStreamInfoFixedSize : fileStreamInfoFixedSize+nameLen])
	}
	return nil
}

// MarshalFileStreamInformationList encodes streams as a chained buffer: each
// entry after the first starts on an 8-byte boundary, the padding is zero, and
// the last entry has NextEntryOffset 0 and no padding after it. The
// NextEntryOffset of each input entry is ignored.
func MarshalFileStreamInformationList(streams []FileStreamInformation) ([]byte, error) {
	var out []byte
	last := 0
	for i := range streams {
		e := streams[i]
		e.NextEntryOffset = 0
		b, err := e.Marshal()
		if err != nil {
			return nil, err
		}
		start := (len(out) + 7) &^ 7
		if i > 0 {
			binary.LittleEndian.PutUint32(out[last:], uint32(start-last))
		}
		out = append(out, make([]byte, start-len(out))...)
		out = append(out, b...)
		last = start
	}
	return out, nil
}

// ParseFileStreamInformationList decodes a chained buffer of
// FILE_STREAM_INFORMATION entries.
func ParseFileStreamInformationList(buf []byte) []FileStreamInformation {
	var out []FileStreamInformation
	for pos := 0; pos+fileStreamInfoFixedSize <= len(buf); {
		var e FileStreamInformation
		if err := e.Unmarshal(buf[pos:]); err != nil {
			break
		}
		out = append(out, e)
		if pos = nextEntry(buf, pos, fileStreamInfoFixedSize, 4); pos < 0 {
			break
		}
	}
	return out
}
