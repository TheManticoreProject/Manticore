package dsrepl

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// ds_repl_neighborw_blob_header_size is the size, in bytes, of the fixed header
// that precedes the variable-length data region.
const ds_repl_neighborw_blob_header_size = 128

// DS_REPL_NEIGHBORW_BLOB is a representation of a tuple from the repsFrom or
// repsTo abstract attribute of an NC replica. This structure, retrieved using an
// LDAP search method, is an alternative representation of DS_REPL_NEIGHBORW,
// retrieved using the IDL_DRSGetReplInfo RPC method.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/58877402-4bf1-4b40-b405-fd9d11691752
type DS_REPL_NEIGHBORW_BLOB struct {
	// NamingContext is the NC to which this replication state data pertains (oszNamingContext). nil if NULL.
	NamingContext *string
	// SourceDsaDN is the DN of the nTDSDSA object of the source server (oszSourceDsaDN). nil if NULL.
	SourceDsaDN *string
	// SourceDsaAddress is the transport-specific network address of the source server (oszSourceDsaAddress). nil if NULL.
	SourceDsaAddress *string
	// AsyncIntersiteTransportDN is the DN of the interSiteTransport object, or nil for RPC/IP (oszAsyncIntersiteTransportDN).
	AsyncIntersiteTransportDN *string
	// ReplicaFlags is a set of DS_REPL_NBR_* flags (dwReplicaFlags).
	ReplicaFlags uint32
	// Reserved is reserved for future use (dwReserved).
	Reserved uint32
	// NamingContextObjGuid is the objectGUID of the NC (uuidNamingContextObjGuid).
	NamingContextObjGuid guid.GUID
	// SourceDsaObjGuid is the objectGUID of the source nTDSDSA object (uuidSourceDsaObjGuid).
	SourceDsaObjGuid guid.GUID
	// SourceDsaInvocationID is the invocationId used by the source server (uuidSourceDsaInvocationID).
	SourceDsaInvocationID guid.GUID
	// AsyncIntersiteTransportObjGuid is the objectGUID of the intersite transport object (uuidAsyncIntersiteTransportObjGuid).
	AsyncIntersiteTransportObjGuid guid.GUID
	// LastObjChangeSynced is the USN of the last object update received (usnLastObjChangeSynced).
	LastObjChangeSynced int64
	// AttributeFilter is the LastObjChangeSynced value at the end of the last complete cycle (usnAttributeFilter).
	AttributeFilter int64
	// LastSyncSuccess is the time the last successful replication cycle completed (ftimeLastSyncSuccess).
	LastSyncSuccess msdtyp.FILETIME
	// LastSyncAttempt is the time of the last replication attempt (ftimeLastSyncAttempt).
	LastSyncAttempt msdtyp.FILETIME
	// LastSyncResult is the Windows error code of the last replication attempt (dwLastSyncResult).
	LastSyncResult uint32
	// NumConsecutiveSyncFailures is the number of failed replication attempts since the last success (cNumConsecutiveSyncFailures).
	NumConsecutiveSyncFailures uint32
}

// NewDS_REPL_NEIGHBORW_BLOB creates a new, empty DS_REPL_NEIGHBORW_BLOB structure.
func NewDS_REPL_NEIGHBORW_BLOB() *DS_REPL_NEIGHBORW_BLOB {
	return &DS_REPL_NEIGHBORW_BLOB{}
}

// Unmarshal parses a DS_REPL_NEIGHBORW_BLOB structure from a byte slice.
//
// Parameters:
// - data: A byte slice containing the structure (header followed by its data region).
//
// Returns:
// - The number of bytes consumed.
// - An error if the unmarshalling fails.
func (b *DS_REPL_NEIGHBORW_BLOB) Unmarshal(data []byte) (int, error) {
	if len(data) < ds_repl_neighborw_blob_header_size {
		return 0, fmt.Errorf("data is too short to unmarshal DS_REPL_NEIGHBORW_BLOB (expected at least %d bytes, got %d)", ds_repl_neighborw_blob_header_size, len(data))
	}

	oszNamingContext := binary.LittleEndian.Uint32(data[0:4])
	oszSourceDsaDN := binary.LittleEndian.Uint32(data[4:8])
	oszSourceDsaAddress := binary.LittleEndian.Uint32(data[8:12])
	oszAsyncIntersiteTransportDN := binary.LittleEndian.Uint32(data[12:16])

	b.ReplicaFlags = binary.LittleEndian.Uint32(data[16:20])
	b.Reserved = binary.LittleEndian.Uint32(data[20:24])

	b.NamingContextObjGuid.FromRawBytes(data[24:40])
	b.SourceDsaObjGuid.FromRawBytes(data[40:56])
	b.SourceDsaInvocationID.FromRawBytes(data[56:72])
	b.AsyncIntersiteTransportObjGuid.FromRawBytes(data[72:88])

	b.LastObjChangeSynced = int64(binary.LittleEndian.Uint64(data[88:96]))
	b.AttributeFilter = int64(binary.LittleEndian.Uint64(data[96:104]))

	if _, err := b.LastSyncSuccess.Unmarshal(data[104:112]); err != nil {
		return 0, err
	}
	if _, err := b.LastSyncAttempt.Unmarshal(data[112:120]); err != nil {
		return 0, err
	}

	b.LastSyncResult = binary.LittleEndian.Uint32(data[120:124])
	b.NumConsecutiveSyncFailures = binary.LittleEndian.Uint32(data[124:128])

	var err error
	if b.NamingContext, err = readOffsetString(data, oszNamingContext); err != nil {
		return 0, err
	}
	if b.SourceDsaDN, err = readOffsetString(data, oszSourceDsaDN); err != nil {
		return 0, err
	}
	if b.SourceDsaAddress, err = readOffsetString(data, oszSourceDsaAddress); err != nil {
		return 0, err
	}
	if b.AsyncIntersiteTransportDN, err = readOffsetString(data, oszAsyncIntersiteTransportDN); err != nil {
		return 0, err
	}

	return len(data), nil
}

// Marshal serializes the DS_REPL_NEIGHBORW_BLOB structure into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled structure.
// - An error if the marshalling fails.
func (b *DS_REPL_NEIGHBORW_BLOB) Marshal() ([]byte, error) {
	data := newDataRegion(ds_repl_neighborw_blob_header_size)

	header := make([]byte, ds_repl_neighborw_blob_header_size)

	binary.LittleEndian.PutUint32(header[0:4], data.addString(b.NamingContext))
	binary.LittleEndian.PutUint32(header[4:8], data.addString(b.SourceDsaDN))
	binary.LittleEndian.PutUint32(header[8:12], data.addString(b.SourceDsaAddress))
	binary.LittleEndian.PutUint32(header[12:16], data.addString(b.AsyncIntersiteTransportDN))

	binary.LittleEndian.PutUint32(header[16:20], b.ReplicaFlags)
	binary.LittleEndian.PutUint32(header[20:24], b.Reserved)

	copy(header[24:40], b.NamingContextObjGuid.ToBytes())
	copy(header[40:56], b.SourceDsaObjGuid.ToBytes())
	copy(header[56:72], b.SourceDsaInvocationID.ToBytes())
	copy(header[72:88], b.AsyncIntersiteTransportObjGuid.ToBytes())

	binary.LittleEndian.PutUint64(header[88:96], uint64(b.LastObjChangeSynced))
	binary.LittleEndian.PutUint64(header[96:104], uint64(b.AttributeFilter))

	lastSyncSuccess, err := b.LastSyncSuccess.Marshal()
	if err != nil {
		return nil, err
	}
	copy(header[104:112], lastSyncSuccess)

	lastSyncAttempt, err := b.LastSyncAttempt.Marshal()
	if err != nil {
		return nil, err
	}
	copy(header[112:120], lastSyncAttempt)

	binary.LittleEndian.PutUint32(header[120:124], b.LastSyncResult)
	binary.LittleEndian.PutUint32(header[124:128], b.NumConsecutiveSyncFailures)

	return append(header, data.bytes()...), nil
}

// String returns a string representation of the DS_REPL_NEIGHBORW_BLOB structure.
func (b *DS_REPL_NEIGHBORW_BLOB) String() string {
	return fmt.Sprintf("DS_REPL_NEIGHBORW_BLOB: NamingContext=%s, SourceDsaDN=%s, ReplicaFlags=0x%08x", describeOszString(b.NamingContext), describeOszString(b.SourceDsaDN), b.ReplicaFlags)
}

// Describe prints the DS_REPL_NEIGHBORW_BLOB structure to the console.
//
// Parameters:
// - indent: The number of levels to indent the output.
func (b *DS_REPL_NEIGHBORW_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mDS_REPL_NEIGHBORW_BLOB\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mNamingContext\x1b[0m: %s\n", indentPrompt, describeOszString(b.NamingContext))
	fmt.Printf("%s │ \x1b[93mSourceDsaDN\x1b[0m: %s\n", indentPrompt, describeOszString(b.SourceDsaDN))
	fmt.Printf("%s │ \x1b[93mSourceDsaAddress\x1b[0m: %s\n", indentPrompt, describeOszString(b.SourceDsaAddress))
	fmt.Printf("%s │ \x1b[93mAsyncIntersiteTransportDN\x1b[0m: %s\n", indentPrompt, describeOszString(b.AsyncIntersiteTransportDN))
	fmt.Printf("%s │ \x1b[93mReplicaFlags\x1b[0m: 0x%08x\n", indentPrompt, b.ReplicaFlags)
	fmt.Printf("%s │ \x1b[93mReserved\x1b[0m: 0x%08x\n", indentPrompt, b.Reserved)
	fmt.Printf("%s │ \x1b[93mNamingContextObjGuid\x1b[0m: %s\n", indentPrompt, b.NamingContextObjGuid.ToFormatD())
	fmt.Printf("%s │ \x1b[93mSourceDsaObjGuid\x1b[0m: %s\n", indentPrompt, b.SourceDsaObjGuid.ToFormatD())
	fmt.Printf("%s │ \x1b[93mSourceDsaInvocationID\x1b[0m: %s\n", indentPrompt, b.SourceDsaInvocationID.ToFormatD())
	fmt.Printf("%s │ \x1b[93mAsyncIntersiteTransportObjGuid\x1b[0m: %s\n", indentPrompt, b.AsyncIntersiteTransportObjGuid.ToFormatD())
	fmt.Printf("%s │ \x1b[93mLastObjChangeSynced\x1b[0m: %d\n", indentPrompt, b.LastObjChangeSynced)
	fmt.Printf("%s │ \x1b[93mAttributeFilter\x1b[0m: %d\n", indentPrompt, b.AttributeFilter)
	fmt.Printf("%s │ \x1b[93mLastSyncSuccess\x1b[0m: %s\n", indentPrompt, b.LastSyncSuccess.String())
	fmt.Printf("%s │ \x1b[93mLastSyncAttempt\x1b[0m: %s\n", indentPrompt, b.LastSyncAttempt.String())
	fmt.Printf("%s │ \x1b[93mLastSyncResult\x1b[0m: %d\n", indentPrompt, b.LastSyncResult)
	fmt.Printf("%s │ \x1b[93mNumConsecutiveSyncFailures\x1b[0m: %d\n", indentPrompt, b.NumConsecutiveSyncFailures)
	fmt.Printf("%s └───\n", indentPrompt)
}
