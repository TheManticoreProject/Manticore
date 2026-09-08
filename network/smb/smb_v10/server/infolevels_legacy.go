package server

import (
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/informationlevels"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// The pre-NT information levels, which a client that did not negotiate
// CAP_NT_FIND uses ([MS-CIFS] sections 2.2.8.1 to 2.2.8.3).
const (
	smbInfoStandard        = 0x0001
	smbInfoQueryEaSize     = 0x0002
	smbInfoQueryEasFromLst = 0x0003
	smbInfoIsNameValid     = 0x0006

	smbInfoAllocation = 0x0001
	smbInfoVolume     = 0x0002

	smbQueryFileStreamInfo      = 0x0109
	smbQueryFileCompressionInfo = 0x010B
)

// findLevelChains reports whether a find level's entries are linked by a leading
// NextEntryOffset.
//
// The NT levels are: each entry begins with the offset of the next and the last
// carries zero. The pre-NT levels are not — SMB_INFO_STANDARD begins with a date
// — and are packed back to back, with the client counting them by SearchCount.
//
// Getting this wrong is not a near miss. Clearing a "final NextEntryOffset" in a
// buffer of pre-NT entries would write zeroes over the first entry's creation
// date and access date, and the client would read a valid-looking listing with
// corrupted timestamps.
func findLevelChains(level uint16) bool {
	switch level {
	case smbInfoStandard, smbInfoQueryEaSize, smbInfoQueryEasFromLst:
		return false
	}
	return true
}

// infoStandardOf fills SMB_INFO_STANDARD from what the backend reported.
//
// The two sizes are truncated to 32 bits, which is what the level's fields are.
// A file larger than 4 GiB cannot be described by this level at all, so a client
// that needs to see one has to ask at an NT level; that is a property of the
// format rather than something this can work around.
func infoStandardOf(attr FileAttr) informationlevels.SMB_INFO_STANDARD {
	standard := informationlevels.SMB_INFO_STANDARD{}
	standard.Creationdate, standard.Creationtime = dosDateTimeOf(attr.Created)
	standard.Lastaccessdate, standard.Lastaccesstime = dosDateTimeOf(attr.Accessed)
	standard.Lastwritedate, standard.Lastwritetime = dosDateTimeOf(attr.Modified)
	standard.Filedatasize = types.ULONG(uint32(attr.Size))
	standard.Allocationsize = types.ULONG(uint32(attr.AllocationSize))
	standard.Attributes.SetAttributes(legacyAttributesFor(attr))
	return standard
}

// encodeLegacyFindEntry renders one directory entry in a pre-NT find level.
//
// These entries carry the name length in a single byte and no NextEntryOffset, so
// they are packed one after another. The resume key is not included: it is present
// only when the client sets SMB_FIND_RETURN_RESUME_KEYS, and the searches here are
// continued by search identifier rather than by key.
//
// Parameters:
//   - level: the find level requested
//   - attr: the entry to render
//   - unicode: whether the message declared Unicode
//
// Returns:
//   - The encoded entry, and whether the level is served
func encodeLegacyFindEntry(level uint16, attr FileAttr, unicode bool) ([]byte, bool) {
	switch level {
	case smbInfoStandard, smbInfoQueryEaSize, smbInfoQueryEasFromLst:
	default:
		return nil, false
	}

	standard := infoStandardOf(attr)
	entry, err := standard.Marshal()
	if err != nil {
		return nil, false
	}

	// The two EA-bearing levels are SMB_INFO_STANDARD followed by a 4-byte EaSize.
	// Nothing here carries extended attributes, so it is zero — which is the
	// documented way to say a file has none, not a placeholder.
	if level == smbInfoQueryEaSize || level == smbInfoQueryEasFromLst {
		entry = append(entry, 0x00, 0x00, 0x00, 0x00)
	}

	name := encodeWireString(attr.Name, unicode)
	// FileNameLength is one byte, so a name that cannot be described is dropped
	// rather than truncated: a truncated name is a different file.
	if len(name) > 0xFF {
		return nil, false
	}
	entry = append(entry, byte(len(name)))
	entry = append(entry, name...)

	// The name is null-terminated in this level, in the width the message
	// declared.
	if unicode {
		entry = append(entry, 0x00, 0x00)
	} else {
		entry = append(entry, 0x00)
	}
	return entry, true
}

// encodeLegacyFileInformation renders a file in a pre-NT query level.
//
// Parameters:
//   - level: the query level requested
//   - attr: what the backend reported
//
// Returns:
//   - The encoded structure, and whether the level is served
func encodeLegacyFileInformation(level uint16, attr FileAttr) ([]byte, bool) {
	switch level {
	case smbInfoStandard:
		standard := infoStandardOf(attr)
		encoded, err := standard.Marshal()
		return encoded, err == nil

	case smbInfoQueryEaSize:
		standard := infoStandardOf(attr)
		encoded, err := standard.Marshal()
		if err != nil {
			return nil, false
		}
		// EaSize(4), zero for storage with no extended attributes.
		return append(encoded, 0x00, 0x00, 0x00, 0x00), true

	case smbInfoIsNameValid:
		// [MS-CIFS] section 2.2.8.3.5: "No parameters or data are returned on this
		// InformationLevel request. An error is returned if the syntax of the name
		// is incorrect." The name has already been through resolvePath by the time
		// this runs, so reaching here means it is valid and success with no data
		// is the whole answer.
		return []byte{}, true

	case smbQueryFileStreamInfo:
		return encodeStreamInformation(attr), true

	case smbQueryFileCompressionInfo:
		compression := informationlevels.SMB_QUERY_FILE_COMRESSION_INFO{
			Compressedfilesize: types.LARGE_INTEGER{QuadPart: uint64(attr.Size)},
			// COMPRESSION_FORMAT_NONE. Nothing here compresses, and claiming a
			// format would have a client compute sizes from a unit shift that
			// does not describe the storage.
			Compressionformat: types.USHORT(0),
		}
		encoded, err := compression.Marshal()
		return encoded, err == nil
	}

	return nil, false
}

// encodeStreamInformation describes an entry's data streams.
//
// A file has exactly one: the unnamed data stream, which is named "::$DATA" in
// this structure. A directory has none, and an empty buffer is how that is
// reported — a client asking a directory for its streams gets a successful answer
// with no entries rather than an error.
func encodeStreamInformation(attr FileAttr) []byte {
	if attr.IsDir {
		return []byte{}
	}

	// The stream name is UTF-16LE regardless of what the message declared: this
	// structure is the native FILE_STREAM_INFORMATION, and the SMB level is a
	// pass-through of it.
	name := utf16.EncodeUTF16LE(`::$DATA`)

	stream := informationlevels.SMB_QUERY_FILE_STREAM_INFO{
		// One entry, so NextEntryOffset is zero and the chain ends here.
		Nextentryoffset:      types.ULONG(0),
		Streamnamelength:     types.ULONG(len(name)),
		Streamsize:           types.LARGE_INTEGER{QuadPart: uint64(attr.Size)},
		Streamallocationsize: types.LARGE_INTEGER{QuadPart: uint64(attr.AllocationSize)},
		Streamname:           []types.UCHAR(name),
	}
	encoded, err := stream.Marshal()
	if err != nil {
		return []byte{}
	}
	return encoded
}

// encodeLegacyVolumeInformation renders a volume in a pre-NT query level.
//
// Parameters:
//   - level: the query level requested
//   - volume: what the backend reported
//   - unicode: whether the message declared Unicode
//
// Returns:
//   - The encoded structure, and whether the level is served
func encodeLegacyVolumeInformation(level uint16, volume VolumeInfo, unicode bool) ([]byte, bool) {
	switch level {
	case smbInfoAllocation:
		unit := int64(volume.SectorsPerAllocationUnit) * int64(volume.BytesPerSector)
		if unit <= 0 {
			unit = defaultAllocationUnit
		}

		allocation := informationlevels.SMB_INFO_ALLOCATION{
			// idFileSystem is a file system identifier; the volume's serial number
			// is the only stable one available.
			Idfilesystem:   types.ULONG(volume.SerialNumber),
			Csectorunit:    types.ULONG(volume.SectorsPerAllocationUnit),
			Cunit:          types.ULONG(uint32(volume.TotalBytes / unit)),
			Cunitavailable: types.ULONG(uint32(volume.FreeBytes / unit)),
			Cbsector:       types.USHORT(volume.BytesPerSector),
		}
		encoded, err := allocation.Marshal()
		return encoded, err == nil

	case smbInfoVolume:
		label := encodeWireString(volume.Label, unicode)

		// A label longer than the one-byte count can describe is cut, because the
		// count is what the client reads the label's extent from: sending more
		// than it can be told about would leave the excess to be parsed as
		// whatever field a client expected next.
		if len(label) > 0xFF {
			label = label[:0xFF]
		}

		// Ccharcount is not set here: SMB_INFO_VOLUME.Marshal derives it from the
		// label's byte length, which is deliberate and worth knowing about.
		// [MS-CIFS] calls the field "the number of characters in the VolumeLabel
		// field", and the two readings differ under Unicode — but they coincide
		// in OEM, which is what a client old enough to ask for this level speaks,
		// and the shared structure's choice is not this handler's to override.
		info := informationlevels.SMB_INFO_VOLUME{
			Ulvolserialnbr: types.ULONG(volume.SerialNumber),
			Volumelabel:    []types.UCHAR(label),
		}
		encoded, err := info.Marshal()
		return encoded, err == nil
	}

	return nil, false
}

// defaultAllocationUnit is assumed when a backend reports no geometry, so a
// division by it cannot be a division by zero.
const defaultAllocationUnit = 4096
