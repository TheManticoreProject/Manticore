package dsrepl

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
)

// ds_repl_attr_meta_data_blob_header_size is the size, in bytes, of the fixed
// header that precedes the variable-length data region.
const ds_repl_attr_meta_data_blob_header_size = 52

// DS_REPL_ATTR_META_DATA_BLOB is a representation of a stamp variable (of type
// AttributeStamp) of an attribute. This structure, retrieved using an LDAP search
// method, is an alternative representation of DS_REPL_ATTR_META_DATA_2, retrieved
// using the IDL_DRSGetReplInfo RPC method.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/63fabec7-cb4b-47ed-ab45-9a20e4397d55
type DS_REPL_ATTR_META_DATA_BLOB struct {
	// AttributeName is the LDAP display name of the attribute (oszAttributeName). nil if NULL.
	AttributeName *string
	// Version is the dwVersion of the attribute's AttributeStamp (dwVersion).
	Version uint32
	// LastOriginatingChange is the timeChanged of the AttributeStamp (ftimeLastOriginatingChange).
	LastOriginatingChange data_structures.FILETIME
	// LastOriginatingDsaInvocationID is the uuidOriginating of the AttributeStamp (uuidLastOriginatingDsaInvocationID).
	LastOriginatingDsaInvocationID guid.GUID
	// OriginatingChange is the usnOriginating of the AttributeStamp (usnOriginatingChange).
	OriginatingChange int64
	// LocalChange is the USN on the destination server of the last applied change (usnLocalChange).
	LocalChange int64
	// LastOriginatingDsaDN is the DN of the nTDSDSA object that originated the last replication (oszLastOriginatingDsaDN). nil if NULL.
	LastOriginatingDsaDN *string
}

// NewDS_REPL_ATTR_META_DATA_BLOB creates a new, empty DS_REPL_ATTR_META_DATA_BLOB structure.
func NewDS_REPL_ATTR_META_DATA_BLOB() *DS_REPL_ATTR_META_DATA_BLOB {
	return &DS_REPL_ATTR_META_DATA_BLOB{}
}

// Unmarshal parses a DS_REPL_ATTR_META_DATA_BLOB structure from a byte slice.
//
// Parameters:
// - data: A byte slice containing the structure (header followed by its data region).
//
// Returns:
// - The number of bytes consumed.
// - An error if the unmarshalling fails.
func (b *DS_REPL_ATTR_META_DATA_BLOB) Unmarshal(data []byte) (int, error) {
	if len(data) < ds_repl_attr_meta_data_blob_header_size {
		return 0, fmt.Errorf("data is too short to unmarshal DS_REPL_ATTR_META_DATA_BLOB (expected at least %d bytes, got %d)", ds_repl_attr_meta_data_blob_header_size, len(data))
	}

	oszAttributeName := binary.LittleEndian.Uint32(data[0:4])

	b.Version = binary.LittleEndian.Uint32(data[4:8])

	if _, err := b.LastOriginatingChange.Unmarshal(data[8:16]); err != nil {
		return 0, err
	}

	b.LastOriginatingDsaInvocationID.FromRawBytes(data[16:32])

	b.OriginatingChange = int64(binary.LittleEndian.Uint64(data[32:40]))
	b.LocalChange = int64(binary.LittleEndian.Uint64(data[40:48]))

	oszLastOriginatingDsaDN := binary.LittleEndian.Uint32(data[48:52])

	var err error
	if b.AttributeName, err = readOffsetString(data, oszAttributeName); err != nil {
		return 0, err
	}
	if b.LastOriginatingDsaDN, err = readOffsetString(data, oszLastOriginatingDsaDN); err != nil {
		return 0, err
	}

	return len(data), nil
}

// Marshal serializes the DS_REPL_ATTR_META_DATA_BLOB structure into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled structure.
// - An error if the marshalling fails.
func (b *DS_REPL_ATTR_META_DATA_BLOB) Marshal() ([]byte, error) {
	data := newDataRegion(ds_repl_attr_meta_data_blob_header_size)

	header := make([]byte, ds_repl_attr_meta_data_blob_header_size)

	binary.LittleEndian.PutUint32(header[0:4], data.addString(b.AttributeName))

	binary.LittleEndian.PutUint32(header[4:8], b.Version)

	lastOriginatingChange, err := b.LastOriginatingChange.Marshal()
	if err != nil {
		return nil, err
	}
	copy(header[8:16], lastOriginatingChange)

	copy(header[16:32], b.LastOriginatingDsaInvocationID.ToBytes())

	binary.LittleEndian.PutUint64(header[32:40], uint64(b.OriginatingChange))
	binary.LittleEndian.PutUint64(header[40:48], uint64(b.LocalChange))

	binary.LittleEndian.PutUint32(header[48:52], data.addString(b.LastOriginatingDsaDN))

	return append(header, data.bytes()...), nil
}

// String returns a string representation of the DS_REPL_ATTR_META_DATA_BLOB structure.
func (b *DS_REPL_ATTR_META_DATA_BLOB) String() string {
	return fmt.Sprintf("DS_REPL_ATTR_META_DATA_BLOB: AttributeName=%s, Version=%d, LastOriginatingDsaDN=%s", describeOszString(b.AttributeName), b.Version, describeOszString(b.LastOriginatingDsaDN))
}

// Describe prints the DS_REPL_ATTR_META_DATA_BLOB structure to the console.
//
// Parameters:
// - indent: The number of levels to indent the output.
func (b *DS_REPL_ATTR_META_DATA_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mDS_REPL_ATTR_META_DATA_BLOB\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mAttributeName\x1b[0m: %s\n", indentPrompt, describeOszString(b.AttributeName))
	fmt.Printf("%s │ \x1b[93mVersion\x1b[0m: %d\n", indentPrompt, b.Version)
	fmt.Printf("%s │ \x1b[93mLastOriginatingChange\x1b[0m: %s\n", indentPrompt, b.LastOriginatingChange.String())
	fmt.Printf("%s │ \x1b[93mLastOriginatingDsaInvocationID\x1b[0m: %s\n", indentPrompt, b.LastOriginatingDsaInvocationID.ToFormatD())
	fmt.Printf("%s │ \x1b[93mOriginatingChange\x1b[0m: %d\n", indentPrompt, b.OriginatingChange)
	fmt.Printf("%s │ \x1b[93mLocalChange\x1b[0m: %d\n", indentPrompt, b.LocalChange)
	fmt.Printf("%s │ \x1b[93mLastOriginatingDsaDN\x1b[0m: %s\n", indentPrompt, describeOszString(b.LastOriginatingDsaDN))
	fmt.Printf("%s └───\n", indentPrompt)
}
