package informationlevels

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_QUERY_FILE_COMRESSION_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/1211daed-3d93-42ae-bf22-c8554d7bbe97
type SMB_QUERY_FILE_COMRESSION_INFO struct {
	// CompressedFileSize: (8 bytes): A 64-bit signed integer that contains the size,
	// in bytes, of the compressed file. This value MUST be greater than or equal to
	// 0x0000000000000000.
	Compressedfilesize types.LARGE_INTEGER
	// CompressionFormat: (2 bytes): A 16-bit unsigned integer that contains the
	// compression format. The actual compression operation associated with each of
	// these compression format values is implementation-dependent. An implementation
	// can associate any local compression algorithm with the values described in the
	// following table, because the compressed data does not travel across the wire in
	// the context of this transaction. The following compression formats are valid
	// only for NTFS.
	Compressionformat types.USHORT
	// CompressionUnitShift: (1 byte): An 8-bit unsigned integer that contains the
	// compression unit shift that is the number of bits by which to left-shift a 1 bit
	// to arrive at the compression unit size. The compression unit size is the number
	// of bytes in a compression unit, that is, the number of bytes to be compressed.
	// This value is implementation-defined.
	Compressionunitshift types.UCHAR
	// ChunkShift: (1 byte): An 8-bit unsigned integer that contains the compression
	// chunk size in bytes in log 2 format. The chunk size is the number of bytes that
	// the operating system's implementation of the Lempel-Ziv compression algorithm
	// tries to compress at one time. This value is implementation-defined.
	Chunkshift types.UCHAR
	// ClusterShift: (1 byte): An 8-bit unsigned integer that specifies, in log 2
	// format, the amount of space that MUST be saved by compression to successfully
	// compress a compression unit. If that amount of space is not saved by
	// compression, the data in that compression unit MUST be stored uncompressed. Each
	// successfully compressed compression unit MUST occupy at least one cluster that
	// is less in bytes than an uncompressed compression unit. Therefore, the cluster
	// shift is the number of bits by which to left shift a 1 bit to arrive at the size
	// of a cluster. This value is implementation-defined.
	Clustershift types.UCHAR
	// Reserved: (3 bytes): A 24-bit reserved value. This field SHOULD be set to
	// 0x000000 and MUST be ignored.
	Reserved [3]types.UCHAR
}

// Marshal serializes the SMB_QUERY_FILE_COMRESSION_INFO into a byte slice.
//
// This method marshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The marshalled data follows the specific format required for this information level.
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_QUERY_FILE_COMRESSION_INFO) Marshal() ([]byte, error) {
	marshalled_struct := make([]byte, 0, 16)

	// CompressedFileSize (8 bytes, LARGE_INTEGER).
	buf8 := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf8, uint64(s.Compressedfilesize.QuadPart))
	marshalled_struct = append(marshalled_struct, buf8...)

	// CompressionFormat (2 bytes).
	buf2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(s.Compressionformat))
	marshalled_struct = append(marshalled_struct, buf2...)

	// CompressionUnitShift, ChunkShift, ClusterShift (1 byte each).
	marshalled_struct = append(marshalled_struct, byte(s.Compressionunitshift), byte(s.Chunkshift), byte(s.Clustershift))

	// Reserved (3 bytes).
	marshalled_struct = append(marshalled_struct, byte(s.Reserved[0]), byte(s.Reserved[1]), byte(s.Reserved[2]))

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FILE_COMRESSION_INFO structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FILE_COMRESSION_INFO structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FILE_COMRESSION_INFO) Unmarshal(data []byte) (int, error) {
	// CompressedFileSize(8) + CompressionFormat(2) + 3x shift(1) + Reserved(3) = 16 bytes.
	if len(data) < 16 {
		return 0, fmt.Errorf("data too short for SMB_QUERY_FILE_COMPRESSION_INFO (need 16 bytes, have %d)", len(data))
	}
	s.Compressedfilesize.QuadPart = binary.LittleEndian.Uint64(data[0:8])
	s.Compressionformat = types.USHORT(binary.LittleEndian.Uint16(data[8:10]))
	s.Compressionunitshift = types.UCHAR(data[10])
	s.Chunkshift = types.UCHAR(data[11])
	s.Clustershift = types.UCHAR(data[12])
	s.Reserved[0] = types.UCHAR(data[13])
	s.Reserved[1] = types.UCHAR(data[14])
	s.Reserved[2] = types.UCHAR(data[15])

	return 16, nil
}
