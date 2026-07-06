package dsrepl

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// ds_repl_kcc_dsa_failurew_blob_header_size is the size, in bytes, of the fixed
// header that precedes the variable-length data region.
const ds_repl_kcc_dsa_failurew_blob_header_size = 36

// DS_REPL_KCC_DSA_FAILUREW_BLOB is a representation of a tuple from the
// kCCFailedConnections or kCCFailedLinks variables of a DC. This structure,
// retrieved using an LDAP search method, is an alternative representation of
// DS_REPL_KCC_DSA_FAILUREW, retrieved using the IDL_DRSGetReplInfo RPC method.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/93083f06-d18a-407b-b906-630ed214a7ff
type DS_REPL_KCC_DSA_FAILUREW_BLOB struct {
	// DsaDN is the DN of the nTDSDSA object of the source server (oszDsaDN). nil if NULL.
	DsaDN *string
	// DsaObjGuid is the objectGUID of the object represented by DsaDN (uuidDsaObjGuid).
	DsaObjGuid guid.GUID
	// FirstFailure is the time the first failure occurred (ftimeFirstFailure).
	FirstFailure msdtyp.FILETIME
	// NumFailures is the number of consecutive failures since the last success (cNumFailures).
	NumFailures uint32
	// LastResult is the error code associated with the most recent failure (dwLastResult).
	LastResult uint32
}

// NewDS_REPL_KCC_DSA_FAILUREW_BLOB creates a new, empty DS_REPL_KCC_DSA_FAILUREW_BLOB structure.
func NewDS_REPL_KCC_DSA_FAILUREW_BLOB() *DS_REPL_KCC_DSA_FAILUREW_BLOB {
	return &DS_REPL_KCC_DSA_FAILUREW_BLOB{}
}

// Unmarshal parses a DS_REPL_KCC_DSA_FAILUREW_BLOB structure from a byte slice.
//
// Parameters:
// - data: A byte slice containing the structure (header followed by its data region).
//
// Returns:
// - The number of bytes consumed.
// - An error if the unmarshalling fails.
func (b *DS_REPL_KCC_DSA_FAILUREW_BLOB) Unmarshal(data []byte) (int, error) {
	if len(data) < ds_repl_kcc_dsa_failurew_blob_header_size {
		return 0, fmt.Errorf("data is too short to unmarshal DS_REPL_KCC_DSA_FAILUREW_BLOB (expected at least %d bytes, got %d)", ds_repl_kcc_dsa_failurew_blob_header_size, len(data))
	}

	oszDsaDN := binary.LittleEndian.Uint32(data[0:4])

	b.DsaObjGuid.FromRawBytes(data[4:20])

	if _, err := b.FirstFailure.Unmarshal(data[20:28]); err != nil {
		return 0, err
	}

	b.NumFailures = binary.LittleEndian.Uint32(data[28:32])
	b.LastResult = binary.LittleEndian.Uint32(data[32:36])

	var err error
	if b.DsaDN, err = readOffsetString(data, oszDsaDN); err != nil {
		return 0, err
	}

	return len(data), nil
}

// Marshal serializes the DS_REPL_KCC_DSA_FAILUREW_BLOB structure into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled structure.
// - An error if the marshalling fails.
func (b *DS_REPL_KCC_DSA_FAILUREW_BLOB) Marshal() ([]byte, error) {
	data := newDataRegion(ds_repl_kcc_dsa_failurew_blob_header_size)

	header := make([]byte, ds_repl_kcc_dsa_failurew_blob_header_size)

	binary.LittleEndian.PutUint32(header[0:4], data.addString(b.DsaDN))

	copy(header[4:20], b.DsaObjGuid.ToBytes())

	firstFailure, err := b.FirstFailure.Marshal()
	if err != nil {
		return nil, err
	}
	copy(header[20:28], firstFailure)

	binary.LittleEndian.PutUint32(header[28:32], b.NumFailures)
	binary.LittleEndian.PutUint32(header[32:36], b.LastResult)

	return append(header, data.bytes()...), nil
}

// String returns a string representation of the DS_REPL_KCC_DSA_FAILUREW_BLOB structure.
func (b *DS_REPL_KCC_DSA_FAILUREW_BLOB) String() string {
	return fmt.Sprintf("DS_REPL_KCC_DSA_FAILUREW_BLOB: DsaDN=%s, NumFailures=%d, LastResult=%d", describeOszString(b.DsaDN), b.NumFailures, b.LastResult)
}

// Describe prints the DS_REPL_KCC_DSA_FAILUREW_BLOB structure to the console.
//
// Parameters:
// - indent: The number of levels to indent the output.
func (b *DS_REPL_KCC_DSA_FAILUREW_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mDS_REPL_KCC_DSA_FAILUREW_BLOB\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mDsaDN\x1b[0m: %s\n", indentPrompt, describeOszString(b.DsaDN))
	fmt.Printf("%s │ \x1b[93mDsaObjGuid\x1b[0m: %s\n", indentPrompt, b.DsaObjGuid.ToFormatD())
	fmt.Printf("%s │ \x1b[93mFirstFailure\x1b[0m: %s\n", indentPrompt, b.FirstFailure.String())
	fmt.Printf("%s │ \x1b[93mNumFailures\x1b[0m: %d\n", indentPrompt, b.NumFailures)
	fmt.Printf("%s │ \x1b[93mLastResult\x1b[0m: %d\n", indentPrompt, b.LastResult)
	fmt.Printf("%s └───\n", indentPrompt)
}
