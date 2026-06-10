package client

import (
	"time"

	"github.com/TheManticoreProject/Manticore/windows/filesystem"
)

// fileBothDirectoryInformation is the FILE_INFORMATION_CLASS used for directory
// enumeration; its FILE_BOTH_DIR_INFORMATION entries (MS-FSCC 2.4.8) carry the
// name, attributes, size, and timestamps needed for FileInfo. (The create-time
// access/share/disposition/options values live in windows/fileflags.)
const fileBothDirectoryInformation uint8 = 0x03

// parseBothDirectoryInfo decodes a buffer of consecutive FILE_BOTH_DIR_INFORMATION
// entries (MS-FSCC 2.4.8), as returned by SMB2 QUERY_DIRECTORY, into FileInfo
// values. The "." and ".." pseudo-entries are skipped. Decoding is delegated to
// the typed windows/filesystem parser; this maps each entry's FILETIME fields to
// time.Time and projects the subset FileInfo exposes.
func parseBothDirectoryInfo(buf []byte) []FileInfo {
	var out []FileInfo
	for _, e := range filesystem.ParseFileBothDirectoryInformationList(buf) {
		if e.FileName == "." || e.FileName == ".." {
			continue
		}
		out = append(out, FileInfo{
			Name:           e.FileName,
			FileAttributes: e.FileAttributes,
			Size:           uint64(e.EndOfFile),
			AllocationSize: uint64(e.AllocationSize),
			CreationTime:   filetimeToTime(e.CreationTime),
			LastAccessTime: filetimeToTime(e.LastAccessTime),
			LastWriteTime:  filetimeToTime(e.LastWriteTime),
			ChangeTime:     filetimeToTime(e.ChangeTime),
		})
	}
	return out
}

// filetimeToTime converts a Windows FILETIME (100ns ticks since 1601-01-01 UTC)
// to a time.Time. A zero (or sub-epoch) FILETIME maps to the zero time.
func filetimeToTime(ft uint64) time.Time {
	const epochDiff = 116444736000000000 // 100ns intervals between 1601-01-01 and 1970-01-01
	if ft == 0 || ft < epochDiff {
		return time.Time{}
	}
	return time.Unix(0, int64(ft-epochDiff)*100).UTC()
}
