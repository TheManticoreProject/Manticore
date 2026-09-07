package keycredentiallink

import (
	"fmt"
	"sort"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"
)

// KEYCREDENTIALLINK_BLOB represents a key credential link blob structure used for authentication and authorization.
//
// The KEYCREDENTIALLINK_BLOB structure is a representation of a single credential stored as a series of values.
// This structure is stored as the binary portion of the msDS-KeyCredentialLink DN-Binary attribute (section 3.1.1.5.3.1.1.6).
// The structure contains a Version field followed by an array of KEYCREDENTIALLINK_ENTRY structures (section 2.2.20.3).
// The KEYCREDENTIALLINK_ENTRY structure MUST be sorted by their Identifier fields in increasing order.
//
// All keys MUST contain KeyID, KeyMaterial, and KeyUsage entries.
// Keys SHOULD contain KeyHash,KeyApproximateLastLogonTimeStamp, and KeyCreationTime entries.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/f3f01e95-6d0c-4fe6-8b43-d585167658fa
type KEYCREDENTIALLINK_BLOB struct {
	// A KeyCredentialLinkVersion object representing the version of the key credential.
	Version version.KeyCredentialLinkVersion
	// A byte slice containing the entries of the key credential link blob.
	Entries []KEYCREDENTIALLINK_ENTRY
}

// NewKEYCREDENTIALLINK_BLOB creates a new KEYCREDENTIALLINK_BLOB structure.
//
// Parameters:
// - version: A KeyCredentialLinkVersion object representing the version of the key credential.
// - entries: A slice of KEYCREDENTIALLINK_ENTRY objects representing the entries of the key credential link blob.
//
// Returns:
// - A pointer to a KEYCREDENTIALLINK_BLOB object.
func NewKEYCREDENTIALLINK_BLOB(version version.KeyCredentialLinkVersion, entries []KEYCREDENTIALLINK_ENTRY) *KEYCREDENTIALLINK_BLOB {
	return &KEYCREDENTIALLINK_BLOB{Version: version, Entries: entries}
}

// Unmarshal unmarshals the KEYCREDENTIALLINK_BLOB structure from a byte slice.
//
// Parameters:
// - data: A byte slice containing the data of the key credential link blob.
//
// Returns:
// - An error if the unmarshalling fails, otherwise nil.
func (k *KEYCREDENTIALLINK_BLOB) Unmarshal(data []byte) (int, error) {
	bytesRead := 0

	// The version is a fixed four-byte field, and the entry loop below already
	// refuses a truncated entry header, so the version needs the same check
	// rather than slicing a buffer that may be shorter than the field.
	if len(data) < 4 {
		return bytesRead, fmt.Errorf("malformed KeyCredentialLink: insufficient bytes for the version (expected at least 4, got %d)", len(data))
	}

	if _, err := k.Version.Unmarshal(data[bytesRead : bytesRead+4]); err != nil {
		return bytesRead, fmt.Errorf("failed to unmarshal KeyCredentialLink version: %w", err)
	}
	bytesRead += 4

	k.Entries = make([]KEYCREDENTIALLINK_ENTRY, 0)
	for bytesRead < len(data) {
		// Ensure there are enough bytes for the entry header (2 bytes for length + 1 byte for type).
		if len(data[bytesRead:]) < 3 {
			// If there are remaining bytes but not enough for a full header, the data is malformed.
			return bytesRead, fmt.Errorf("malformed KeyCredentialLink: insufficient bytes for entry header (expected at least 3, got %d remaining)", len(data[bytesRead:]))
		}

		entry := KEYCREDENTIALLINK_ENTRY{}
		bytesReadEntry, err := entry.Unmarshal(data[bytesRead:])
		if err != nil {
			return bytesRead, err
		}
		bytesRead += bytesReadEntry
		k.Entries = append(k.Entries, entry)
	}

	return bytesRead, nil
}

// Marshal marshals the KEYCREDENTIALLINK_BLOB structure to a byte slice.
//
// Parameters:
// - None
//
// Returns:
// - A byte slice containing the marshalled KEYCREDENTIALLINK_BLOB structure.
func (k *KEYCREDENTIALLINK_BLOB) Marshal() ([]byte, error) {
	marshalledData := make([]byte, 0)

	versionBytes, err := k.Version.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, versionBytes...)

	for _, entry := range k.Entries {
		entryBytes, err := entry.Marshal()
		if err != nil {
			return nil, err
		}
		marshalledData = append(marshalledData, entryBytes...)
	}

	return marshalledData, nil
}

// String returns a string representation of the KEYCREDENTIALLINK_BLOB structure.
//
// Parameters:
// - None
//
// Returns:
// - A string representation of the KEYCREDENTIALLINK_BLOB structure.
func (k *KEYCREDENTIALLINK_BLOB) String() string {
	return fmt.Sprintf("KEYCREDENTIALLINK_BLOB: Version=%s, Entries=%v", k.Version.String(), k.Entries)
}

// Describe prints the KEYCREDENTIALLINK_BLOB structure to the console.
//
// Parameters:
// - indent: The number of spaces to indent the output.
func (k *KEYCREDENTIALLINK_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mKEYCREDENTIALLINK_BLOB\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mVersion\x1b[0m: %s\n", indentPrompt, k.Version.String())
	fmt.Printf("%s │ \x1b[93mEntries\x1b[0m: (%d entries)\n", indentPrompt, len(k.Entries))
	indent += 1
	for i, entry := range k.Entries {
		fmt.Printf("%s │  │ \x1b[93mEntry #%d\x1b[0m:\n", indentPrompt, i+1)
		entry.Describe(indent + 2)
		fmt.Printf("%s │  │  └───\n", indentPrompt)
	}
	fmt.Printf("%s │  └───\n", indentPrompt)
	indent -= 1
	fmt.Printf("%s └───\n", indentPrompt)
}

// SortEntriesByType sorts the KEYCREDENTIALLINK_BLOB entries in ascending order by their Identifier.Value.
//
// Parameters:
// - None
//
// Returns:
// - None
//
// Note:
// The function sorts the entries by their Identifier.Value in ascending order.
func (k *KEYCREDENTIALLINK_BLOB) SortEntriesByType() {
	if len(k.Entries) <= 1 {
		return
	}

	sort.Slice(k.Entries, func(i, j int) bool {
		return k.Entries[i].Identifier < k.Entries[j].Identifier
	})
}

func (b *KEYCREDENTIALLINK_BLOB) RemoveEntryByType(entryType KEYCREDENTIALLINK_ENTRY_IDENTIFIER) {
	for i, entry := range b.Entries {
		if entry.Identifier == entryType {
			b.Entries = append(b.Entries[:i], b.Entries[i+1:]...)
			break
		}
	}
}
