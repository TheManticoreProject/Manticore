package server

import (
	"encoding/binary"
	"time"

	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// Information levels are the shapes a client can ask for a file, a directory
// listing or a volume to be described in. Each is a fixed wire layout, so the
// encoders below build the bytes directly rather than through a struct: what
// matters is the offsets, and writing them out makes the offsets the thing being
// reviewed.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/f0a5ba31-1d84-4bcb-a0f1-6e4f0e3b30f6
const (
	// Find information levels.
	smbFindFileDirectoryInfo     = 0x0101
	smbFindFileFullDirectoryInfo = 0x0102
	smbFindFileNamesInfo         = 0x0103
	smbFindFileBothDirectoryInfo = 0x0104

	// Query file information levels.
	smbQueryFileBasicInfo    = 0x0101
	smbQueryFileStandardInfo = 0x0102
	smbQueryFileEAInfo       = 0x0103
	smbQueryFileNameInfo     = 0x0104
	smbQueryFileAllInfo      = 0x0107
	smbQueryFileAltNameInfo  = 0x0108

	// Set file information levels.
	smbSetFileBasicInfo       = 0x0101
	smbSetFileDispositionInfo = 0x0102
	smbSetFileAllocationInfo  = 0x0103
	smbSetFileEndOfFileInfo   = 0x0104

	// Query file-system information levels.
	smbQueryFsVolumeInfo    = 0x0102
	smbQueryFsSizeInfo      = 0x0103
	smbQueryFsDeviceInfo    = 0x0104
	smbQueryFsAttributeInfo = 0x0105
)

// bothDirectoryInfoFixedSize is the fixed part of an
// SMB_FIND_FILE_BOTH_DIRECTORY_INFO entry, before its variable-length name.
const bothDirectoryInfoFixedSize = 94

// filetimeOf renders a time as the 64-bit Windows FILETIME the wire carries. A
// zero time renders as zero rather than as the year 1601, which is what a client
// displays for a converted zero.
func filetimeOf(when time.Time) uint64 {
	if when.IsZero() {
		return 0
	}
	const unixEpochIn100ns = 116444736000000000
	return uint64(when.UTC().UnixNano()/100 + unixEpochIn100ns)
}

// supportedFindLevel reports whether a directory listing can be returned in a
// level. Refusing an unsupported level is better than returning the nearest one:
// a client parses what it asked for, so a different shape is read as corruption.
func supportedFindLevel(level uint16) bool {
	switch level {
	case smbFindFileDirectoryInfo, smbFindFileFullDirectoryInfo,
		smbFindFileNamesInfo, smbFindFileBothDirectoryInfo:
		return true
	}
	return false
}

// encodeFindEntries takes up to count entries from a search and renders them,
// stopping early if the next entry would not fit in the client's budget.
//
// Stopping short is not a failure: the search keeps its position, and the client
// asks again. What must not happen is a truncated entry, which a client would
// parse as a corrupt one.
//
// Returns:
//   - The encoded entries
//   - How many were encoded, which the response reports as its count
func encodeFindEntries(search *Search, count, budget int, unicode bool) ([]byte, int) {
	if count <= 0 {
		count = len(search.Entries)
	}
	if budget <= 0 {
		budget = maxTrans2Payload
	}

	encoded := []byte{}
	returned := 0

	for returned < count && !search.exhausted() {
		entry := search.Entries[search.Position]
		rendered := encodeFindEntry(search.InformationLevel, entry.Attr, unicode)

		// The entry must fit whole.
		if len(encoded)+len(rendered) > budget {
			break
		}

		encoded = append(encoded, rendered...)
		search.Position++
		returned++
	}

	// The last entry's NextEntryOffset is zero, which is how a client knows to
	// stop walking the buffer. Patching it afterwards is simpler than knowing in
	// advance which entry will be last, since the budget decides that.
	if returned > 0 && search.InformationLevel != smbFindFileNamesInfo {
		zeroLastNextEntryOffset(encoded, search.InformationLevel)
	}

	return encoded, returned
}

// zeroLastNextEntryOffset walks the encoded entries and clears the final
// NextEntryOffset.
func zeroLastNextEntryOffset(encoded []byte, level uint16) {
	position := 0
	for position+4 <= len(encoded) {
		next := int(binary.LittleEndian.Uint32(encoded[position : position+4]))
		if next == 0 || position+next > len(encoded) {
			// Already terminated, or the walk would leave the buffer.
			binary.LittleEndian.PutUint32(encoded[position:position+4], 0)
			return
		}
		if position+next == len(encoded) {
			binary.LittleEndian.PutUint32(encoded[position:position+4], 0)
			return
		}
		position += next
	}
}

// encodeFindEntry renders one directory entry in the requested level.
//
// Names go out in the encoding the request declared, and the lengths that describe
// them are byte counts either way. Writing them as OEM regardless is not a
// cosmetic difference: a client that negotiated Unicode reads the buffer as UTF-16
// whatever the server put there, so an OEM name comes out as half as many
// characters of nonsense.
func encodeFindEntry(level uint16, attr FileAttr, unicode bool) []byte {
	name := encodeWireString(attr.Name, unicode)

	switch level {
	case smbFindFileNamesInfo:
		// NextEntryOffset(4) FileIndex(4) FileNameLength(4) FileName(variable).
		entry := make([]byte, 12)
		binary.LittleEndian.PutUint32(entry[8:12], uint32(len(name)))
		entry = append(entry, name...)
		entry = padTo4(entry)
		binary.LittleEndian.PutUint32(entry[0:4], uint32(len(entry)))
		return entry

	case smbFindFileDirectoryInfo:
		// The fixed part is the first 64 bytes of the both-directory layout,
		// without the EA size or the short name.
		entry := make([]byte, 64)
		writeDirectoryInfoCommon(entry, attr, len(name))
		entry = append(entry, name...)
		entry = padTo4(entry)
		binary.LittleEndian.PutUint32(entry[0:4], uint32(len(entry)))
		return entry

	case smbFindFileFullDirectoryInfo:
		// As above, plus a 4-byte EA size.
		entry := make([]byte, 68)
		writeDirectoryInfoCommon(entry, attr, len(name))
		entry = append(entry, name...)
		entry = padTo4(entry)
		binary.LittleEndian.PutUint32(entry[0:4], uint32(len(entry)))
		return entry

	default:
		// SMB_FIND_FILE_BOTH_DIRECTORY_INFO, which is what a Windows client asks
		// for and what this repository's client parses.
		entry := make([]byte, bothDirectoryInfoFixedSize)
		writeDirectoryInfoCommon(entry, attr, len(name))
		// ShortNameLength(1) at 68, Reserved(1) at 69, ShortName(24) at 70. The
		// short name is left empty: this server has no 8.3 alias to report, and a
		// client uses the long name when the short one is absent.
		entry = append(entry, name...)
		entry = padTo4(entry)
		binary.LittleEndian.PutUint32(entry[0:4], uint32(len(entry)))
		return entry
	}
}

// writeDirectoryInfoCommon fills the part every directory-info level shares:
// NextEntryOffset(4) FileIndex(4) four timestamps(8 each) EndOfFile(8)
// AllocationSize(8) ExtFileAttributes(4) FileNameLength(4).
func writeDirectoryInfoCommon(entry []byte, attr FileAttr, nameLength int) {
	binary.LittleEndian.PutUint64(entry[8:16], filetimeOf(attr.Created))
	binary.LittleEndian.PutUint64(entry[16:24], filetimeOf(attr.Accessed))
	binary.LittleEndian.PutUint64(entry[24:32], filetimeOf(attr.Modified))
	binary.LittleEndian.PutUint64(entry[32:40], filetimeOf(attr.Changed))
	binary.LittleEndian.PutUint64(entry[40:48], uint64(attr.Size))
	binary.LittleEndian.PutUint64(entry[48:56], uint64(attr.AllocationSize))
	binary.LittleEndian.PutUint32(entry[56:60], attributesFor(attr))
	binary.LittleEndian.PutUint32(entry[60:64], uint32(nameLength))
}

// padTo4 rounds a buffer up to a 4-byte boundary, which each entry is aligned to
// so the next one starts aligned.
func padTo4(buffer []byte) []byte {
	if remainder := len(buffer) % 4; remainder != 0 {
		buffer = append(buffer, make([]byte, 4-remainder)...)
	}
	return buffer
}

// encodeFileInformation renders a file in a query level, or reports that the level
// is not served.
func encodeFileInformation(level uint16, attr FileAttr, path string, unicode bool) ([]byte, bool) {
	switch level {
	case smbQueryFileBasicInfo:
		// Four timestamps(8 each) ExtFileAttributes(4) Reserved(4).
		info := make([]byte, 40)
		binary.LittleEndian.PutUint64(info[0:8], filetimeOf(attr.Created))
		binary.LittleEndian.PutUint64(info[8:16], filetimeOf(attr.Accessed))
		binary.LittleEndian.PutUint64(info[16:24], filetimeOf(attr.Modified))
		binary.LittleEndian.PutUint64(info[24:32], filetimeOf(attr.Changed))
		binary.LittleEndian.PutUint32(info[32:36], attributesFor(attr))
		return info, true

	case smbQueryFileStandardInfo:
		// AllocationSize(8) EndOfFile(8) NumberOfLinks(4) DeletePending(1)
		// Directory(1) Reserved(2).
		info := make([]byte, 24)
		binary.LittleEndian.PutUint64(info[0:8], uint64(attr.AllocationSize))
		binary.LittleEndian.PutUint64(info[8:16], uint64(attr.Size))
		binary.LittleEndian.PutUint32(info[16:20], 1)
		if attr.IsDir {
			info[21] = 1
		}
		return info, true

	case smbQueryFileEAInfo:
		// EaSize(4). Nothing here carries extended attributes, so it is zero.
		return make([]byte, 4), true

	case smbQueryFileNameInfo, smbQueryFileAltNameInfo:
		// FileNameLength(4) FileName(variable).
		name := encodeWireString(pathForClient(path), unicode)
		info := make([]byte, 4)
		binary.LittleEndian.PutUint32(info[0:4], uint32(len(name)))
		return append(info, name...), true

	case smbQueryFileAllInfo:
		// The basic and standard blocks together, then EaSize and the name.
		info := make([]byte, 0, 104)
		basic, _ := encodeFileInformation(smbQueryFileBasicInfo, attr, path, unicode)
		standard, _ := encodeFileInformation(smbQueryFileStandardInfo, attr, path, unicode)
		info = append(info, basic...)
		info = append(info, standard...)
		info = append(info, make([]byte, 4)...) // EaSize
		name := encodeWireString(pathForClient(path), unicode)
		lengthField := make([]byte, 4)
		binary.LittleEndian.PutUint32(lengthField, uint32(len(name)))
		info = append(info, lengthField...)
		return append(info, name...), true
	}

	return nil, false
}

// pathForClient renders a resolved path back in the form a client sent it:
// backslash separated and rooted.
func pathForClient(path string) string {
	rendered := "\\"
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			rendered += "\\"
			continue
		}
		rendered += string(path[i])
	}
	return rendered
}

// encodeVolumeInformation renders a volume in a query level, or reports that the
// level is not served.
func encodeVolumeInformation(level uint16, volume VolumeInfo, unicode bool) ([]byte, bool) {
	// A level at or above the pass-through base names a native information class
	// rather than an SMB one ([MS-SMB] section 2.2.2.3.5). A client that asks for
	// free space asks this way, so refusing the range means refusing the question
	// most often asked about a volume.
	if level >= smbInfoPassthrough {
		return encodeNativeVolumeInformation(level-smbInfoPassthrough, volume, unicode)
	}

	switch level {
	case smbQueryFsVolumeInfo:
		// VolumeCreationTime(8) SerialNumber(4) VolumeLabelSize(4) Reserved(2)
		// VolumeLabel(variable).
		label := encodeWireString(volume.Label, unicode)
		info := make([]byte, 18)
		binary.LittleEndian.PutUint32(info[8:12], volume.SerialNumber)
		binary.LittleEndian.PutUint32(info[12:16], uint32(len(label)))
		return append(info, label...), true

	case smbQueryFsSizeInfo:
		// TotalAllocationUnits(8) TotalFreeAllocationUnits(8)
		// SectorsPerAllocationUnit(4) BytesPerSector(4).
		unit := int64(volume.SectorsPerAllocationUnit) * int64(volume.BytesPerSector)
		if unit <= 0 {
			unit = 4096
		}
		info := make([]byte, 24)
		binary.LittleEndian.PutUint64(info[0:8], uint64(volume.TotalBytes/unit))
		binary.LittleEndian.PutUint64(info[8:16], uint64(volume.FreeBytes/unit))
		binary.LittleEndian.PutUint32(info[16:20], volume.SectorsPerAllocationUnit)
		binary.LittleEndian.PutUint32(info[20:24], volume.BytesPerSector)
		return info, true

	case smbQueryFsDeviceInfo:
		// DeviceType(4) DeviceCharacteristics(4). FILE_DEVICE_DISK with no
		// special characteristics.
		info := make([]byte, 8)
		binary.LittleEndian.PutUint32(info[0:4], 0x00000007)
		return info, true

	case smbQueryFsAttributeInfo:
		// FileSystemAttributes(4) MaxFileNameLengthInBytes(4)
		// LengthOfFileSystemName(4) FileSystemName(variable).
		//
		// Only case-preserving names are claimed. Claiming a capability the
		// storage does not have — unicode-on-disk, compression, quotas — is worse
		// than claiming none, because a client then uses it.
		name := encodeWireString(volume.FileSystemName, unicode)
		info := make([]byte, 12)
		binary.LittleEndian.PutUint32(info[0:4], 0x00000002) // FILE_CASE_PRESERVED_NAMES
		binary.LittleEndian.PutUint32(info[4:8], MaxPathComponentLength)
		binary.LittleEndian.PutUint32(info[8:12], uint32(len(name)))
		return append(info, name...), true
	}

	return nil, false
}

// applyFileInformation applies a set level to a path, reporting whether the level
// is served and what the backend made of it.
func applyFileInformation(fs FileSystem, path string, level uint16, data []byte, open *Open) (bool, error) {
	switch level {
	case smbSetFileBasicInfo:
		if len(data) < 36 {
			return true, ErrAccessDenied
		}
		attr := FileAttr{
			Created:  timeFromFiletime(binary.LittleEndian.Uint64(data[0:8])),
			Accessed: timeFromFiletime(binary.LittleEndian.Uint64(data[8:16])),
			Modified: timeFromFiletime(binary.LittleEndian.Uint64(data[16:24])),
			Changed:  timeFromFiletime(binary.LittleEndian.Uint64(data[24:32])),
			ReadOnly: binary.LittleEndian.Uint32(data[32:36])&fileflags.FILE_ATTRIBUTE_READONLY != 0,
		}
		// A zero timestamp means "leave this one alone", so the mask selects only
		// the ones the client actually supplied.
		mask := AttrMask{
			ReadOnly: true,
			Created:  !attr.Created.IsZero(),
			Accessed: !attr.Accessed.IsZero(),
			Modified: !attr.Modified.IsZero(),
			Changed:  !attr.Changed.IsZero(),
		}
		return true, fs.SetAttr(path, attr, mask)

	case smbSetFileDispositionInfo:
		if len(data) < 1 {
			return true, ErrAccessDenied
		}
		// Delete-on-close is a property of the handle, not of the file, so it is
		// recorded rather than acted on now.
		if open == nil {
			return true, ErrAccessDenied
		}
		open.DeleteOnClose = data[0] != 0
		return true, nil

	case smbSetFileEndOfFileInfo, smbSetFileAllocationInfo:
		if len(data) < 8 {
			return true, ErrAccessDenied
		}
		size := int64(binary.LittleEndian.Uint64(data[0:8]))
		if size < 0 {
			return true, ErrAccessDenied
		}
		// The allocation level asks for space to be reserved rather than for the
		// length to change, and nothing here distinguishes the two, so only the
		// end-of-file level resizes.
		if level == smbSetFileAllocationInfo {
			return true, nil
		}
		return true, fs.SetAttr(path, FileAttr{Size: size}, AttrMask{Size: true})
	}

	return false, nil
}

// timeFromFiletime converts a Windows FILETIME to a time. Zero and the
// leave-alone sentinel both render as the zero time, which the mask above reads as
// "not supplied".
func timeFromFiletime(filetime uint64) time.Time {
	const unixEpochIn100ns = 116444736000000000
	if filetime == 0 || filetime == 0xFFFFFFFFFFFFFFFF || filetime < unixEpochIn100ns {
		return time.Time{}
	}
	return time.Unix(0, int64(filetime-unixEpochIn100ns)*100).UTC()
}

// smbInfoPassthrough is the base of the pass-through information levels: a level
// of smbInfoPassthrough + N carries the native information class N.
const smbInfoPassthrough = 0x03E8

// Native file-system information classes, as they arrive through the pass-through
// range ([MS-FSCC] section 2.5).
const (
	fileFsVolumeInformation    = 1
	fileFsSizeInformation      = 3
	fileFsDeviceInformation    = 4
	fileFsAttributeInformation = 5
	fileFsFullSizeInformation  = 7
)

// encodeNativeVolumeInformation renders a volume in a native information class.
//
// Most of the classes have the same layout as the SMB level that shadows them, so
// those are answered by the level rather than duplicated. Only the full-size class
// has no SMB equivalent, and it is the one a client actually uses: it reports the
// caller's available space separately from the volume's, which is what a client
// displays as free space.
func encodeNativeVolumeInformation(class uint16, volume VolumeInfo, unicode bool) ([]byte, bool) {
	switch class {
	case fileFsVolumeInformation:
		return encodeVolumeInformation(smbQueryFsVolumeInfo, volume, unicode)
	case fileFsSizeInformation:
		return encodeVolumeInformation(smbQueryFsSizeInfo, volume, unicode)
	case fileFsDeviceInformation:
		return encodeVolumeInformation(smbQueryFsDeviceInfo, volume, unicode)
	case fileFsAttributeInformation:
		return encodeVolumeInformation(smbQueryFsAttributeInfo, volume, unicode)

	case fileFsFullSizeInformation:
		// TotalAllocationUnits(8) CallerAvailableAllocationUnits(8)
		// ActualAvailableAllocationUnits(8) SectorsPerAllocationUnit(4)
		// BytesPerSector(4).
		sectors, bytesPerSector := volume.SectorsPerAllocationUnit, volume.BytesPerSector
		if sectors == 0 {
			sectors = defaultSectorsPerAllocationUnit
		}
		if bytesPerSector == 0 {
			bytesPerSector = defaultBytesPerSector
		}
		unit := int64(sectors) * int64(bytesPerSector)

		info := make([]byte, 32)
		binary.LittleEndian.PutUint64(info[0:8], uint64(volume.TotalBytes/unit))
		// Nothing here imposes a quota, so what the caller may use is what the
		// volume has. Reporting a smaller figure would make a client refuse a write
		// the server would have accepted.
		binary.LittleEndian.PutUint64(info[8:16], uint64(volume.FreeBytes/unit))
		binary.LittleEndian.PutUint64(info[16:24], uint64(volume.FreeBytes/unit))
		binary.LittleEndian.PutUint32(info[24:28], sectors)
		binary.LittleEndian.PutUint32(info[28:32], bytesPerSector)
		return info, true
	}

	return nil, false
}
