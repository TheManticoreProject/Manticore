package server

import (
	"encoding/binary"
	"testing"
)

// TestFsAttributeInfoClaimsUnicodeOnDisk guards the defect: the server reported
// only FILE_CASE_PRESERVED_NAMES, so a client was told the volume could not store
// Unicode names. It can — a UTF-16LE name from a third-party client round-trips
// through the server intact — so withholding the bit made clients avoid names the
// server would have stored correctly.
func TestFsAttributeInfoClaimsUnicodeOnDisk(t *testing.T) {
	volume := VolumeInfo{
		Label:                    "FILES",
		FileSystemName:           "NTFS",
		SerialNumber:             0x11223344,
		BytesPerSector:           512,
		SectorsPerAllocationUnit: 8,
	}

	for _, unicode := range []bool{false, true} {
		info, served := encodeVolumeInformation(smbQueryFsAttributeInfo, volume, unicode)
		if !served {
			t.Fatalf("unicode=%v: SMB_QUERY_FS_ATTRIBUTE_INFO not served", unicode)
		}
		if len(info) < 12 {
			t.Fatalf("unicode=%v: response is %d bytes, want at least 12", unicode, len(info))
		}

		attrs := binary.LittleEndian.Uint32(info[0:4])
		if attrs&fileUnicodeOnDisk == 0 {
			t.Errorf("unicode=%v: FileSystemAttributes = %#08x, missing FILE_UNICODE_ON_DISK", unicode, attrs)
		}
		if attrs&fileCasePreservedNames == 0 {
			t.Errorf("unicode=%v: FileSystemAttributes = %#08x, missing FILE_CASE_PRESERVED_NAMES", unicode, attrs)
		}
		if attrs != serverFileSystemAttributes {
			t.Errorf("unicode=%v: FileSystemAttributes = %#08x, want %#08x", unicode, attrs, serverFileSystemAttributes)
		}

		// MaxFileNameLengthInBytes, then the name length and the name itself.
		if got := binary.LittleEndian.Uint32(info[4:8]); got != MaxPathComponentLength {
			t.Errorf("unicode=%v: MaxFileNameLengthInBytes = %d, want %d", unicode, got, MaxPathComponentLength)
		}
		nameLen := int(binary.LittleEndian.Uint32(info[8:12]))
		if 12+nameLen != len(info) {
			t.Errorf("unicode=%v: declared name length %d does not match the %d trailing bytes", unicode, nameLen, len(info)-12)
		}
	}
}

// TestServerFileSystemAttributesClaimsNothingUnimplemented is the counterweight: it
// fails if a bit is ever added for a feature the server does not implement. Claiming
// one is worse than claiming none, because a client then uses it and the follow-up
// request is refused.
func TestServerFileSystemAttributesClaimsNothingUnimplemented(t *testing.T) {
	const (
		fileCaseSensitiveSearch   uint32 = 0x00000001
		filePersistentACLs        uint32 = 0x00000008
		fileFileCompression       uint32 = 0x00000010
		fileVolumeQuotas          uint32 = 0x00000020
		fileSupportsSparseFiles   uint32 = 0x00000040
		fileSupportsReparsePoints uint32 = 0x00000080
		fileSupportsObjectIDs     uint32 = 0x00010000
		fileSupportsEncryption    uint32 = 0x00020000
		fileNamedStreams          uint32 = 0x00040000
		fileSupportsHardLinks     uint32 = 0x00400000
		fileSupportsExtendedAttrs uint32 = 0x00800000
	)

	unimplemented := map[string]uint32{
		"FILE_CASE_SENSITIVE_SEARCH":        fileCaseSensitiveSearch,
		"FILE_PERSISTENT_ACLS":              filePersistentACLs,
		"FILE_FILE_COMPRESSION":             fileFileCompression,
		"FILE_VOLUME_QUOTAS":                fileVolumeQuotas,
		"FILE_SUPPORTS_SPARSE_FILES":        fileSupportsSparseFiles,
		"FILE_SUPPORTS_REPARSE_POINTS":      fileSupportsReparsePoints,
		"FILE_SUPPORTS_OBJECT_IDS":          fileSupportsObjectIDs,
		"FILE_SUPPORTS_ENCRYPTION":          fileSupportsEncryption,
		"FILE_NAMED_STREAMS":                fileNamedStreams,
		"FILE_SUPPORTS_HARD_LINKS":          fileSupportsHardLinks,
		"FILE_SUPPORTS_EXTENDED_ATTRIBUTES": fileSupportsExtendedAttrs,
	}

	for name, bit := range unimplemented {
		if serverFileSystemAttributes&bit != 0 {
			t.Errorf("serverFileSystemAttributes claims %s, which this server does not implement", name)
		}
	}
}
