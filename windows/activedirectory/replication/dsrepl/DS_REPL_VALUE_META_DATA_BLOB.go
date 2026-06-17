package dsrepl

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
)

// ds_repl_value_meta_data_blob_header_size is the size, in bytes, of the fixed
// header that precedes the variable-length data region.
const ds_repl_value_meta_data_blob_header_size = 80

// DS_REPL_VALUE_META_DATA_BLOB is a representation of a stamp variable (of type
// LinkValueStamp) of a link value. This structure, retrieved using an LDAP search
// method, is an alternative representation of DS_REPL_VALUE_META_DATA_2, retrieved
// using the IDL_DRSGetReplInfo RPC method.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/5d421689-808b-466d-8d9f-106cda2dbcce
type DS_REPL_VALUE_META_DATA_BLOB struct {
	// AttributeName is the LDAP display name of the attribute (oszAttributeName).
	AttributeName string
	// ObjectDn is the DN of the object that this attribute belongs to (oszObjectDn).
	ObjectDn string
	// Data is the attribute replication metadata buffer (cbData / pbData).
	Data []byte
	// Deleted is the timeDeleted of the LinkValueStamp (ftimeDeleted).
	Deleted data_structures.FILETIME
	// Created is the timeCreated of the LinkValueStamp (ftimeCreated).
	Created data_structures.FILETIME
	// Version is the dwVersion of the LinkValueStamp (dwVersion).
	Version uint32
	// LastOriginatingChange is the timeChanged of the LinkValueStamp (ftimeLastOriginatingChange).
	LastOriginatingChange data_structures.FILETIME
	// LastOriginatingDsaInvocationID is the uuidOriginating of the LinkValueStamp (uuidLastOriginatingDsaInvocationID).
	LastOriginatingDsaInvocationID guid.GUID
	// OriginatingChange is the usnOriginating of the LinkValueStamp (usnOriginatingChange).
	OriginatingChange int64
	// LocalChange is the USN on the destination server of the last applied change (usnLocalChange).
	LocalChange int64
	// LastOriginatingDsaDN is the DN of the nTDSDSA object that originated the last replication (oszLastOriginatingDsaDN).
	LastOriginatingDsaDN string
}

// NewDS_REPL_VALUE_META_DATA_BLOB creates a new, empty DS_REPL_VALUE_META_DATA_BLOB structure.
func NewDS_REPL_VALUE_META_DATA_BLOB() *DS_REPL_VALUE_META_DATA_BLOB {
	return &DS_REPL_VALUE_META_DATA_BLOB{}
}

// Unmarshal parses a DS_REPL_VALUE_META_DATA_BLOB structure from a byte slice.
//
// Parameters:
// - data: A byte slice containing the structure (header followed by its data region).
//
// Returns:
// - The number of bytes consumed.
// - An error if the unmarshalling fails.
func (b *DS_REPL_VALUE_META_DATA_BLOB) Unmarshal(data []byte) (int, error) {
	if len(data) < ds_repl_value_meta_data_blob_header_size {
		return 0, fmt.Errorf("data is too short to unmarshal DS_REPL_VALUE_META_DATA_BLOB (expected at least %d bytes, got %d)", ds_repl_value_meta_data_blob_header_size, len(data))
	}

	oszAttributeName := binary.LittleEndian.Uint32(data[0:4])
	oszObjectDn := binary.LittleEndian.Uint32(data[4:8])
	cbData := binary.LittleEndian.Uint32(data[8:12])
	pbData := binary.LittleEndian.Uint32(data[12:16])

	if _, err := b.Deleted.Unmarshal(data[16:24]); err != nil {
		return 0, err
	}
	if _, err := b.Created.Unmarshal(data[24:32]); err != nil {
		return 0, err
	}

	b.Version = binary.LittleEndian.Uint32(data[32:36])

	if _, err := b.LastOriginatingChange.Unmarshal(data[36:44]); err != nil {
		return 0, err
	}

	b.LastOriginatingDsaInvocationID.FromRawBytes(data[44:60])

	b.OriginatingChange = int64(binary.LittleEndian.Uint64(data[60:68]))
	b.LocalChange = int64(binary.LittleEndian.Uint64(data[68:76]))

	oszLastOriginatingDsaDN := binary.LittleEndian.Uint32(data[76:80])

	var err error
	if b.AttributeName, err = readOffsetString(data, oszAttributeName); err != nil {
		return 0, err
	}
	if b.ObjectDn, err = readOffsetString(data, oszObjectDn); err != nil {
		return 0, err
	}
	if b.LastOriginatingDsaDN, err = readOffsetString(data, oszLastOriginatingDsaDN); err != nil {
		return 0, err
	}

	if pbData != 0 && cbData != 0 {
		if int(pbData)+int(cbData) > len(data) {
			return 0, fmt.Errorf("pbData buffer (offset %d, length %d) is out of range (blob is %d bytes)", pbData, cbData, len(data))
		}
		b.Data = make([]byte, cbData)
		copy(b.Data, data[pbData:pbData+cbData])
	} else {
		b.Data = nil
	}

	return len(data), nil
}

// Marshal serializes the DS_REPL_VALUE_META_DATA_BLOB structure into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled structure.
// - An error if the marshalling fails.
func (b *DS_REPL_VALUE_META_DATA_BLOB) Marshal() ([]byte, error) {
	data := newDataRegion(ds_repl_value_meta_data_blob_header_size)

	header := make([]byte, ds_repl_value_meta_data_blob_header_size)

	binary.LittleEndian.PutUint32(header[0:4], data.addString(b.AttributeName))
	binary.LittleEndian.PutUint32(header[4:8], data.addString(b.ObjectDn))
	binary.LittleEndian.PutUint32(header[8:12], uint32(len(b.Data)))
	binary.LittleEndian.PutUint32(header[12:16], data.addBytes(b.Data))

	deleted, err := b.Deleted.Marshal()
	if err != nil {
		return nil, err
	}
	copy(header[16:24], deleted)

	created, err := b.Created.Marshal()
	if err != nil {
		return nil, err
	}
	copy(header[24:32], created)

	binary.LittleEndian.PutUint32(header[32:36], b.Version)

	lastOriginatingChange, err := b.LastOriginatingChange.Marshal()
	if err != nil {
		return nil, err
	}
	copy(header[36:44], lastOriginatingChange)

	copy(header[44:60], b.LastOriginatingDsaInvocationID.ToBytes())

	binary.LittleEndian.PutUint64(header[60:68], uint64(b.OriginatingChange))
	binary.LittleEndian.PutUint64(header[68:76], uint64(b.LocalChange))

	binary.LittleEndian.PutUint32(header[76:80], data.addString(b.LastOriginatingDsaDN))

	return append(header, data.bytes()...), nil
}

// String returns a string representation of the DS_REPL_VALUE_META_DATA_BLOB structure.
func (b *DS_REPL_VALUE_META_DATA_BLOB) String() string {
	return fmt.Sprintf("DS_REPL_VALUE_META_DATA_BLOB: AttributeName=%q, ObjectDn=%q, Version=%d, DataLen=%d", b.AttributeName, b.ObjectDn, b.Version, len(b.Data))
}

// Describe prints the DS_REPL_VALUE_META_DATA_BLOB structure to the console.
//
// Parameters:
// - indent: The number of levels to indent the output.
func (b *DS_REPL_VALUE_META_DATA_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mDS_REPL_VALUE_META_DATA_BLOB\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mAttributeName\x1b[0m: %q\n", indentPrompt, b.AttributeName)
	fmt.Printf("%s │ \x1b[93mObjectDn\x1b[0m: %q\n", indentPrompt, b.ObjectDn)
	fmt.Printf("%s │ \x1b[93mData\x1b[0m: (%d bytes)\n", indentPrompt, len(b.Data))
	fmt.Printf("%s │ \x1b[93mDeleted\x1b[0m: %s\n", indentPrompt, b.Deleted.String())
	fmt.Printf("%s │ \x1b[93mCreated\x1b[0m: %s\n", indentPrompt, b.Created.String())
	fmt.Printf("%s │ \x1b[93mVersion\x1b[0m: %d\n", indentPrompt, b.Version)
	fmt.Printf("%s │ \x1b[93mLastOriginatingChange\x1b[0m: %s\n", indentPrompt, b.LastOriginatingChange.String())
	fmt.Printf("%s │ \x1b[93mLastOriginatingDsaInvocationID\x1b[0m: %s\n", indentPrompt, b.LastOriginatingDsaInvocationID.ToFormatD())
	fmt.Printf("%s │ \x1b[93mOriginatingChange\x1b[0m: %d\n", indentPrompt, b.OriginatingChange)
	fmt.Printf("%s │ \x1b[93mLocalChange\x1b[0m: %d\n", indentPrompt, b.LocalChange)
	fmt.Printf("%s │ \x1b[93mLastOriginatingDsaDN\x1b[0m: %q\n", indentPrompt, b.LastOriginatingDsaDN)
	fmt.Printf("%s └───\n", indentPrompt)
}
