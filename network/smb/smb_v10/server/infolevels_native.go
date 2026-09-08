package server

import (
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
	"github.com/TheManticoreProject/Manticore/windows/filesystem/infoclass"
)

// The pass-through information classes are native structures, so their strings
// are UTF-16LE whatever encoding the message declared.
//
// This is the one place in the server where a name's encoding does not follow
// SMB_FLAGS2_UNICODE: an SMB information level carries an SMB string, but a
// pass-through level carries the [MS-FSCC] structure verbatim, and that structure
// is defined in UTF-16LE. Reading the message's flag here would produce an OEM
// name inside a native structure, which a client parses as UTF-16 regardless.
func nativeName(value string) []byte {
	return utf16.EncodeUTF16LE(value)
}

// encodeNativeFileInformation renders a file in a native FILE_INFORMATION_CLASS,
// for a query in the pass-through range.
//
// The structures come from windows/filesystem rather than being assembled here,
// so the layouts are the ones the rest of the repository already agrees on and
// already tests.
//
// Parameters:
//   - class: the native information class, with the pass-through base removed
//   - attr: what the backend reported about the file
//   - path: the share-relative path, for the classes that name the file
//
// Returns:
//   - The encoded structure, and whether the class is served
func encodeNativeFileInformation(class uint16, attr FileAttr, path string) ([]byte, bool) {
	switch infoclass.FileInformationClass(class) {
	case infoclass.FileBasicInformation:
		basic := nativeBasicInformation(attr)
		encoded, err := basic.Marshal()
		return encoded, err == nil

	case infoclass.FileStandardInformation:
		standard := nativeStandardInformation(attr)
		encoded, err := standard.Marshal()
		return encoded, err == nil

	case infoclass.FileInternalInformation:
		// IndexNumber(8). A client uses it to tell whether two names are the same
		// file. Nothing here keeps a file index, and inventing one would let a
		// client conclude that two unrelated files are hard links to each other,
		// so it is reported as zero — the value for storage with no index.
		return make([]byte, 8), true

	case infoclass.FileEaInformation:
		// EaSize(4). Nothing here carries extended attributes.
		return make([]byte, 4), true

	case infoclass.FileAccessInformation:
		// AccessFlags(4), which is what the open was granted.
		information := make([]byte, 4)
		binary.LittleEndian.PutUint32(information, nativeAccessFlags(attr))
		return information, true

	case infoclass.FilePositionInformation:
		// CurrentByteOffset(8). The server keeps no file pointer — every read and
		// write carries its own offset — so the position is the start.
		return make([]byte, 8), true

	case infoclass.FileNameInformation, infoclass.FileAlternateNameInformation:
		// FileNameLength(4) FileName(variable). The alternate name is the 8.3
		// alias, and there is none here, so the long name answers both: a client
		// falls back to the long name when no alias is offered.
		name := nativeName(nativePath(path))
		information := make([]byte, 4, 4+len(name))
		binary.LittleEndian.PutUint32(information, uint32(len(name)))
		return append(information, name...), true

	case infoclass.FileNetworkOpenInformation:
		open := filesystem.FileNetworkOpenInformation{
			CreationTime:   filetimeOf(attr.Created),
			LastAccessTime: filetimeOf(attr.Accessed),
			LastWriteTime:  filetimeOf(attr.Modified),
			ChangeTime:     filetimeOf(attr.Changed),
			AllocationSize: attr.AllocationSize,
			EndOfFile:      attr.Size,
			FileAttributes: attributesFor(attr),
		}
		encoded, err := open.Marshal()
		return encoded, err == nil

	case infoclass.FileAllInformation:
		all := filesystem.FileAllInformation{
			Basic:    nativeBasicInformation(attr),
			Standard: nativeStandardInformation(attr),
			EaSize:   0,
			// AccessFlags, Mode and AlignmentRequirement describe the open rather
			// than the file. Mode is zero because no write-through or
			// synchronous-IO mode is in force, and the alignment requirement is
			// byte alignment, which is what storage with no device constraint has.
			AccessFlags:          nativeAccessFlags(attr),
			CurrentByteOffset:    0,
			Mode:                 0,
			AlignmentRequirement: 0,
			FileName:             nativePath(path),
		}
		encoded, err := all.Marshal()
		return encoded, err == nil
	}

	return nil, false
}

// nativeBasicInformation fills FILE_BASIC_INFORMATION from what the backend
// reported.
func nativeBasicInformation(attr FileAttr) filesystem.FileBasicInformation {
	return filesystem.FileBasicInformation{
		CreationTime:   filetimeOf(attr.Created),
		LastAccessTime: filetimeOf(attr.Accessed),
		LastWriteTime:  filetimeOf(attr.Modified),
		ChangeTime:     filetimeOf(attr.Changed),
		FileAttributes: attributesFor(attr),
	}
}

// nativeStandardInformation fills FILE_STANDARD_INFORMATION.
//
// NumberOfLinks is one because nothing here creates a hard link, so every file has
// exactly the one name it is reached by.
func nativeStandardInformation(attr FileAttr) filesystem.FileStandardInformation {
	return filesystem.FileStandardInformation{
		AllocationSize: attr.AllocationSize,
		EndOfFile:      attr.Size,
		NumberOfLinks:  1,
		DeletePending:  false,
		Directory:      attr.IsDir,
	}
}

// nativeAccessFlags reports the access a client has to an entry, as the
// FILE_ACCESS_INFORMATION mask.
//
// A read-only entry does not describe write access, for the same reason the
// reflective security provider does not: a client uses this to predict what it
// will be allowed to do, so a mask that disagreed with the handlers would make the
// client wrong.
func nativeAccessFlags(attr FileAttr) uint32 {
	access := uint32(fileflagsGenericRead)
	if !attr.ReadOnly {
		access |= fileflagsGenericWrite
	}
	return access
}

// nativePath renders a share-relative path the way a native structure names one:
// separated by backslashes and rooted, which is what FILE_NAME_INFORMATION
// carries.
func nativePath(path string) string {
	if path == "" {
		return `\`
	}
	rooted := path
	if rooted[0] != '\\' {
		rooted = `\` + rooted
	}
	return rooted
}

// applyNativeFileInformation applies a native FILE_INFORMATION_CLASS to a path,
// for a set in the pass-through range.
//
// Parameters:
//   - fs: the backend to apply it through
//   - path: the share-relative path being changed
//   - class: the native information class, with the pass-through base removed
//   - data: the structure the client sent
//   - open: the handle the set arrived on, or nil for a set by path
//
// Returns:
//   - Whether the class is served, and what the backend made of it
func applyNativeFileInformation(
	fs FileSystem,
	path string,
	class uint16,
	data []byte,
	open *Open,
) (bool, error) {
	switch infoclass.FileInformationClass(class) {
	case infoclass.FileBasicInformation:
		basic := filesystem.FileBasicInformation{}
		if err := basic.Unmarshal(data); err != nil {
			return true, ErrAccessDenied
		}
		attr := FileAttr{
			Created:  timeFromFiletime(basic.CreationTime),
			Accessed: timeFromFiletime(basic.LastAccessTime),
			Modified: timeFromFiletime(basic.LastWriteTime),
			Changed:  timeFromFiletime(basic.ChangeTime),
			ReadOnly: basic.FileAttributes&fileAttributeReadOnly != 0,
		}
		// A zero timestamp means "leave this one alone", which is the same rule
		// the SMB-level basic set follows.
		mask := AttrMask{
			ReadOnly: true,
			Created:  !attr.Created.IsZero(),
			Accessed: !attr.Accessed.IsZero(),
			Modified: !attr.Modified.IsZero(),
			Changed:  !attr.Changed.IsZero(),
		}
		return true, fs.SetAttr(path, attr, mask)

	case infoclass.FileDispositionInformation:
		disposition := filesystem.FileDispositionInformation{}
		if err := disposition.Unmarshal(data); err != nil {
			return true, ErrAccessDenied
		}
		// Delete-on-close is a property of the handle, so a set by path has
		// nowhere to record it.
		if open == nil {
			return true, ErrAccessDenied
		}
		open.DeleteOnClose = disposition.DeletePending
		return true, nil

	case infoclass.FileEndOfFileInformation:
		endOfFile := filesystem.FileEndOfFileInformation{}
		if err := endOfFile.Unmarshal(data); err != nil {
			return true, ErrAccessDenied
		}
		if endOfFile.EndOfFile < 0 {
			return true, ErrAccessDenied
		}
		return true, fs.SetAttr(path, FileAttr{Size: endOfFile.EndOfFile}, AttrMask{Size: true})

	case infoclass.FileAllocationInformation:
		// Reserving space is not changing the length, and nothing here
		// distinguishes the two, so the request is accepted and the length left
		// alone — the same answer the SMB-level allocation set gives.
		if len(data) < 8 {
			return true, ErrAccessDenied
		}
		return true, nil

	case infoclass.FileRenameInformation:
		rename := filesystem.FileRenameInformation{}
		if err := rename.Unmarshal(data); err != nil {
			return true, ErrAccessDenied
		}
		// RootDirectory names a directory handle the new name is relative to.
		// Nothing here can resolve one, and treating a relative name as if it
		// were share-relative would rename the file somewhere the client did not
		// ask for, so it is refused instead.
		if rename.RootDirectory != 0 {
			return true, ErrAccessDenied
		}

		target, err := resolvePath(rename.FileName)
		if err != nil {
			return true, ErrAccessDenied
		}
		return true, fs.Rename(path, target, rename.ReplaceIfExists)
	}

	return false, nil
}

// fileAttributeReadOnly, fileflagsGenericRead and fileflagsGenericWrite are the
// bits the native structures above use, named here so this file does not depend
// on the SMB-level constants for values that belong to [MS-FSCC].
const (
	fileAttributeReadOnly = 0x00000001
	fileflagsGenericRead  = 0x00000001
	fileflagsGenericWrite = 0x00000002
)
