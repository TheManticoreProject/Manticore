package dsrepl

import (
	"encoding/binary"
	"fmt"
	"strings"

	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// ds_repl_queue_statisticsw_blob_size is the fixed size, in bytes, of the
// structure. It has no variable-length data region.
const ds_repl_queue_statisticsw_blob_size = 52

// DS_REPL_QUEUE_STATISTICSW_BLOB contains the statistics related to the
// replicationQueue variable of a DC, returned by reading the
// msDS-ReplQueueStatistics rootDSE attribute.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/aa5923c9-48bf-4182-93ff-6353dd9f0b00
type DS_REPL_QUEUE_STATISTICSW_BLOB struct {
	// CurrentOpStarted is the time the currently running operation started (ftimeCurrentOpStarted).
	CurrentOpStarted msdtyp.FILETIME
	// NumPendingOps is the number of currently pending operations (cNumPendingOps).
	NumPendingOps uint32
	// OldestSync is the time of the oldest synchronization operation (ftimeOldestSync).
	OldestSync msdtyp.FILETIME
	// OldestAdd is the time of the oldest add operation (ftimeOldestAdd).
	OldestAdd msdtyp.FILETIME
	// OldestMod is the time of the oldest modification operation (ftimeOldestMod).
	OldestMod msdtyp.FILETIME
	// OldestDel is the time of the oldest delete operation (ftimeOldestDel).
	OldestDel msdtyp.FILETIME
	// OldestUpdRefs is the time of the oldest reference update operation (ftimeOldestUpdRefs).
	OldestUpdRefs msdtyp.FILETIME
}

// NewDS_REPL_QUEUE_STATISTICSW_BLOB creates a new, empty DS_REPL_QUEUE_STATISTICSW_BLOB structure.
func NewDS_REPL_QUEUE_STATISTICSW_BLOB() *DS_REPL_QUEUE_STATISTICSW_BLOB {
	return &DS_REPL_QUEUE_STATISTICSW_BLOB{}
}

// Unmarshal parses a DS_REPL_QUEUE_STATISTICSW_BLOB structure from a byte slice.
//
// Parameters:
// - data: A byte slice containing the structure.
//
// Returns:
// - The number of bytes consumed.
// - An error if the unmarshalling fails.
func (b *DS_REPL_QUEUE_STATISTICSW_BLOB) Unmarshal(data []byte) (int, error) {
	if len(data) < ds_repl_queue_statisticsw_blob_size {
		return 0, fmt.Errorf("data is too short to unmarshal DS_REPL_QUEUE_STATISTICSW_BLOB (expected at least %d bytes, got %d)", ds_repl_queue_statisticsw_blob_size, len(data))
	}

	if _, err := b.CurrentOpStarted.Unmarshal(data[0:8]); err != nil {
		return 0, err
	}

	b.NumPendingOps = binary.LittleEndian.Uint32(data[8:12])

	if _, err := b.OldestSync.Unmarshal(data[12:20]); err != nil {
		return 0, err
	}
	if _, err := b.OldestAdd.Unmarshal(data[20:28]); err != nil {
		return 0, err
	}
	if _, err := b.OldestMod.Unmarshal(data[28:36]); err != nil {
		return 0, err
	}
	if _, err := b.OldestDel.Unmarshal(data[36:44]); err != nil {
		return 0, err
	}
	if _, err := b.OldestUpdRefs.Unmarshal(data[44:52]); err != nil {
		return 0, err
	}

	return ds_repl_queue_statisticsw_blob_size, nil
}

// Marshal serializes the DS_REPL_QUEUE_STATISTICSW_BLOB structure into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled structure.
// - An error if the marshalling fails.
func (b *DS_REPL_QUEUE_STATISTICSW_BLOB) Marshal() ([]byte, error) {
	header := make([]byte, ds_repl_queue_statisticsw_blob_size)

	currentOpStarted, err := b.CurrentOpStarted.Marshal()
	if err != nil {
		return nil, err
	}
	copy(header[0:8], currentOpStarted)

	binary.LittleEndian.PutUint32(header[8:12], b.NumPendingOps)

	for _, field := range []struct {
		offset int
		ft     *msdtyp.FILETIME
	}{
		{12, &b.OldestSync},
		{20, &b.OldestAdd},
		{28, &b.OldestMod},
		{36, &b.OldestDel},
		{44, &b.OldestUpdRefs},
	} {
		marshalled, err := field.ft.Marshal()
		if err != nil {
			return nil, err
		}
		copy(header[field.offset:field.offset+8], marshalled)
	}

	return header, nil
}

// String returns a string representation of the DS_REPL_QUEUE_STATISTICSW_BLOB structure.
func (b *DS_REPL_QUEUE_STATISTICSW_BLOB) String() string {
	return fmt.Sprintf("DS_REPL_QUEUE_STATISTICSW_BLOB: NumPendingOps=%d, CurrentOpStarted=%s", b.NumPendingOps, b.CurrentOpStarted.String())
}

// Describe prints the DS_REPL_QUEUE_STATISTICSW_BLOB structure to the console.
//
// Parameters:
// - indent: The number of levels to indent the output.
func (b *DS_REPL_QUEUE_STATISTICSW_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mDS_REPL_QUEUE_STATISTICSW_BLOB\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mCurrentOpStarted\x1b[0m: %s\n", indentPrompt, b.CurrentOpStarted.String())
	fmt.Printf("%s │ \x1b[93mNumPendingOps\x1b[0m: %d\n", indentPrompt, b.NumPendingOps)
	fmt.Printf("%s │ \x1b[93mOldestSync\x1b[0m: %s\n", indentPrompt, b.OldestSync.String())
	fmt.Printf("%s │ \x1b[93mOldestAdd\x1b[0m: %s\n", indentPrompt, b.OldestAdd.String())
	fmt.Printf("%s │ \x1b[93mOldestMod\x1b[0m: %s\n", indentPrompt, b.OldestMod.String())
	fmt.Printf("%s │ \x1b[93mOldestDel\x1b[0m: %s\n", indentPrompt, b.OldestDel.String())
	fmt.Printf("%s │ \x1b[93mOldestUpdRefs\x1b[0m: %s\n", indentPrompt, b.OldestUpdRefs.String())
	fmt.Printf("%s └───\n", indentPrompt)
}
