package dsrepl

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
)

// ds_repl_cursor_blob_header_size is the size, in bytes, of the fixed header that
// precedes the variable-length data region.
const ds_repl_cursor_blob_header_size = 36

// DS_REPL_CURSOR_BLOB is the packet representation of the ReplUpToDateVector type
// of an NC replica. This structure, retrieved using an LDAP search method, is an
// alternative representation of DS_REPL_CURSOR_3W, retrieved using the
// IDL_DRSGetReplInfo RPC method.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/42f0e26e-7627-4e68-814c-afc35a2b5f62
type DS_REPL_CURSOR_BLOB struct {
	// SourceDsaInvocationID is the invocationId of the originating server (uuidSourceDsaInvocationID).
	SourceDsaInvocationID guid.GUID
	// AttributeFilter is the maximum USN recorded for the originating server (usnAttributeFilter).
	AttributeFilter int64
	// LastSyncSuccess is the time of the last successful synchronization (fTimeLastSyncSuccess).
	LastSyncSuccess data_structures.FILETIME
	// SourceDsaDN is the DN of the DSA of the source server (oszSourceDsaDN).
	SourceDsaDN string
}

// NewDS_REPL_CURSOR_BLOB creates a new, empty DS_REPL_CURSOR_BLOB structure.
func NewDS_REPL_CURSOR_BLOB() *DS_REPL_CURSOR_BLOB {
	return &DS_REPL_CURSOR_BLOB{}
}

// Unmarshal parses a DS_REPL_CURSOR_BLOB structure from a byte slice.
//
// Parameters:
// - data: A byte slice containing the structure (header followed by its data region).
//
// Returns:
// - The number of bytes consumed.
// - An error if the unmarshalling fails.
func (b *DS_REPL_CURSOR_BLOB) Unmarshal(data []byte) (int, error) {
	if len(data) < ds_repl_cursor_blob_header_size {
		return 0, fmt.Errorf("data is too short to unmarshal DS_REPL_CURSOR_BLOB (expected at least %d bytes, got %d)", ds_repl_cursor_blob_header_size, len(data))
	}

	b.SourceDsaInvocationID.FromRawBytes(data[0:16])

	b.AttributeFilter = int64(binary.LittleEndian.Uint64(data[16:24]))

	if _, err := b.LastSyncSuccess.Unmarshal(data[24:32]); err != nil {
		return 0, err
	}

	oszSourceDsaDN := binary.LittleEndian.Uint32(data[32:36])

	var err error
	if b.SourceDsaDN, err = readOffsetString(data, oszSourceDsaDN); err != nil {
		return 0, err
	}

	return len(data), nil
}

// Marshal serializes the DS_REPL_CURSOR_BLOB structure into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled structure.
// - An error if the marshalling fails.
func (b *DS_REPL_CURSOR_BLOB) Marshal() ([]byte, error) {
	data := newDataRegion(ds_repl_cursor_blob_header_size)

	header := make([]byte, ds_repl_cursor_blob_header_size)

	copy(header[0:16], b.SourceDsaInvocationID.ToBytes())

	binary.LittleEndian.PutUint64(header[16:24], uint64(b.AttributeFilter))

	lastSyncSuccess, err := b.LastSyncSuccess.Marshal()
	if err != nil {
		return nil, err
	}
	copy(header[24:32], lastSyncSuccess)

	binary.LittleEndian.PutUint32(header[32:36], data.addString(b.SourceDsaDN))

	return append(header, data.bytes()...), nil
}

// String returns a string representation of the DS_REPL_CURSOR_BLOB structure.
func (b *DS_REPL_CURSOR_BLOB) String() string {
	return fmt.Sprintf("DS_REPL_CURSOR_BLOB: SourceDsaInvocationID=%s, AttributeFilter=%d, SourceDsaDN=%q", b.SourceDsaInvocationID.ToFormatD(), b.AttributeFilter, b.SourceDsaDN)
}

// Describe prints the DS_REPL_CURSOR_BLOB structure to the console.
//
// Parameters:
// - indent: The number of levels to indent the output.
func (b *DS_REPL_CURSOR_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mDS_REPL_CURSOR_BLOB\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mSourceDsaInvocationID\x1b[0m: %s\n", indentPrompt, b.SourceDsaInvocationID.ToFormatD())
	fmt.Printf("%s │ \x1b[93mAttributeFilter\x1b[0m: %d\n", indentPrompt, b.AttributeFilter)
	fmt.Printf("%s │ \x1b[93mLastSyncSuccess\x1b[0m: %s\n", indentPrompt, b.LastSyncSuccess.String())
	fmt.Printf("%s │ \x1b[93mSourceDsaDN\x1b[0m: %q\n", indentPrompt, b.SourceDsaDN)
	fmt.Printf("%s └───\n", indentPrompt)
}
