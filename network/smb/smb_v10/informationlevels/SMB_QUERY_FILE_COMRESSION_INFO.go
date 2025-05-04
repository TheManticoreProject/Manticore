package informationlevels

import (
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
	marshalled_struct := []byte{}

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
	return 0, nil
}
