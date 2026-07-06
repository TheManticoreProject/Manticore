package dsrepl

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// ds_repl_opw_blob_header_size is the size, in bytes, of the fixed header that
// precedes the variable-length data region.
const ds_repl_opw_blob_header_size = 68

// DS_REPL_OP_TYPE values for the OpType field of DS_REPL_OPW_BLOB.
const (
	DS_REPL_OP_TYPE_SYNC        uint32 = 0
	DS_REPL_OP_TYPE_ADD         uint32 = 1
	DS_REPL_OP_TYPE_DELETE      uint32 = 2
	DS_REPL_OP_TYPE_MODIFY      uint32 = 3
	DS_REPL_OP_TYPE_UPDATE_REFS uint32 = 4
)

// DS_REPL_OPW_BLOB is a representation of a tuple from the replicationQueue
// variable of a DC. This structure, retrieved using an LDAP search method, is an
// alternative representation of DS_REPL_OPW, retrieved using the
// IDL_DRSGetReplInfo RPC method.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/9bdde2d1-1f61-4829-bf73-3a45f85302a3
type DS_REPL_OPW_BLOB struct {
	// Enqueued is the time this operation was added to the queue (ftimeEnqueued).
	Enqueued msdtyp.FILETIME
	// SerialNumber is the identifier of the operation (ulSerialNumber).
	SerialNumber uint32
	// Priority is the priority value of this operation (ulPriority).
	Priority uint32
	// OpType is the type of operation, one of the DS_REPL_OP_TYPE_* values (opType).
	OpType uint32
	// Options is zero or more DRS option bits, interpreted per OpType (ulOptions).
	Options uint32
	// NamingContext is the DN of the NC associated with this operation (oszNamingContext). nil if NULL.
	NamingContext *string
	// DsaDN is the DN of the nTDSDSA object of the remote server (oszDsaDN). nil if NULL.
	DsaDN *string
	// DsaAddress is the transport-specific network address of the remote server (oszDsaAddress). nil if NULL.
	DsaAddress *string
	// NamingContextObjGuid is the objectGUID of the NC identified by NamingContext (uuidNamingContextObjGuid).
	NamingContextObjGuid guid.GUID
	// DsaObjGuid is the objectGUID of the DSA object identified by DsaDN (uuidDsaObjGuid).
	DsaObjGuid guid.GUID
}

// NewDS_REPL_OPW_BLOB creates a new, empty DS_REPL_OPW_BLOB structure.
func NewDS_REPL_OPW_BLOB() *DS_REPL_OPW_BLOB {
	return &DS_REPL_OPW_BLOB{}
}

// Unmarshal parses a DS_REPL_OPW_BLOB structure from a byte slice.
//
// Parameters:
// - data: A byte slice containing the structure (header followed by its data region).
//
// Returns:
// - The number of bytes consumed.
// - An error if the unmarshalling fails.
func (b *DS_REPL_OPW_BLOB) Unmarshal(data []byte) (int, error) {
	if len(data) < ds_repl_opw_blob_header_size {
		return 0, fmt.Errorf("data is too short to unmarshal DS_REPL_OPW_BLOB (expected at least %d bytes, got %d)", ds_repl_opw_blob_header_size, len(data))
	}

	if _, err := b.Enqueued.Unmarshal(data[0:8]); err != nil {
		return 0, err
	}

	b.SerialNumber = binary.LittleEndian.Uint32(data[8:12])
	b.Priority = binary.LittleEndian.Uint32(data[12:16])
	b.OpType = binary.LittleEndian.Uint32(data[16:20])
	b.Options = binary.LittleEndian.Uint32(data[20:24])

	oszNamingContext := binary.LittleEndian.Uint32(data[24:28])
	oszDsaDN := binary.LittleEndian.Uint32(data[28:32])
	oszDsaAddress := binary.LittleEndian.Uint32(data[32:36])

	b.NamingContextObjGuid.FromRawBytes(data[36:52])
	b.DsaObjGuid.FromRawBytes(data[52:68])

	var err error
	if b.NamingContext, err = readOffsetString(data, oszNamingContext); err != nil {
		return 0, err
	}
	if b.DsaDN, err = readOffsetString(data, oszDsaDN); err != nil {
		return 0, err
	}
	if b.DsaAddress, err = readOffsetString(data, oszDsaAddress); err != nil {
		return 0, err
	}

	return len(data), nil
}

// Marshal serializes the DS_REPL_OPW_BLOB structure into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled structure.
// - An error if the marshalling fails.
func (b *DS_REPL_OPW_BLOB) Marshal() ([]byte, error) {
	data := newDataRegion(ds_repl_opw_blob_header_size)

	header := make([]byte, ds_repl_opw_blob_header_size)

	enqueued, err := b.Enqueued.Marshal()
	if err != nil {
		return nil, err
	}
	copy(header[0:8], enqueued)

	binary.LittleEndian.PutUint32(header[8:12], b.SerialNumber)
	binary.LittleEndian.PutUint32(header[12:16], b.Priority)
	binary.LittleEndian.PutUint32(header[16:20], b.OpType)
	binary.LittleEndian.PutUint32(header[20:24], b.Options)

	binary.LittleEndian.PutUint32(header[24:28], data.addString(b.NamingContext))
	binary.LittleEndian.PutUint32(header[28:32], data.addString(b.DsaDN))
	binary.LittleEndian.PutUint32(header[32:36], data.addString(b.DsaAddress))

	copy(header[36:52], b.NamingContextObjGuid.ToBytes())
	copy(header[52:68], b.DsaObjGuid.ToBytes())

	return append(header, data.bytes()...), nil
}

// String returns a string representation of the DS_REPL_OPW_BLOB structure.
func (b *DS_REPL_OPW_BLOB) String() string {
	return fmt.Sprintf("DS_REPL_OPW_BLOB: SerialNumber=%d, OpType=%d, NamingContext=%s, DsaDN=%s", b.SerialNumber, b.OpType, describeOszString(b.NamingContext), describeOszString(b.DsaDN))
}

// Describe prints the DS_REPL_OPW_BLOB structure to the console.
//
// Parameters:
// - indent: The number of levels to indent the output.
func (b *DS_REPL_OPW_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mDS_REPL_OPW_BLOB\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mEnqueued\x1b[0m: %s\n", indentPrompt, b.Enqueued.String())
	fmt.Printf("%s │ \x1b[93mSerialNumber\x1b[0m: %d\n", indentPrompt, b.SerialNumber)
	fmt.Printf("%s │ \x1b[93mPriority\x1b[0m: %d\n", indentPrompt, b.Priority)
	fmt.Printf("%s │ \x1b[93mOpType\x1b[0m: %d\n", indentPrompt, b.OpType)
	fmt.Printf("%s │ \x1b[93mOptions\x1b[0m: 0x%08x\n", indentPrompt, b.Options)
	fmt.Printf("%s │ \x1b[93mNamingContext\x1b[0m: %s\n", indentPrompt, describeOszString(b.NamingContext))
	fmt.Printf("%s │ \x1b[93mDsaDN\x1b[0m: %s\n", indentPrompt, describeOszString(b.DsaDN))
	fmt.Printf("%s │ \x1b[93mDsaAddress\x1b[0m: %s\n", indentPrompt, describeOszString(b.DsaAddress))
	fmt.Printf("%s │ \x1b[93mNamingContextObjGuid\x1b[0m: %s\n", indentPrompt, b.NamingContextObjGuid.ToFormatD())
	fmt.Printf("%s │ \x1b[93mDsaObjGuid\x1b[0m: %s\n", indentPrompt, b.DsaObjGuid.ToFormatD())
	fmt.Printf("%s └───\n", indentPrompt)
}
