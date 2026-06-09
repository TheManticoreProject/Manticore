// Package fileflags holds Windows file-related bit-flag values: the access mask,
// share access, create disposition, create options, and file attributes.
//
// Although they are best known as the parameters of a Win32 CreateFile /
// NtCreateFile call and the SMB1 NT_CREATE_ANDX / SMB2 CREATE requests, several
// are used well beyond creating a file: the access mask appears in security
// descriptors and the SMB2 TREE_CONNECT MaximalAccess response, and the file
// attributes appear in directory enumeration and FileBasicInformation queries.
// They are shared by the SMB engines, the generic SMB client, and the DCE/RPC
// named-pipe transport.
//
// Combine them with bitwise OR, e.g.:
//
//	DesiredAccess:     fileflags.GENERIC_READ | fileflags.FILE_READ_ATTRIBUTES
//	ShareAccess:       fileflags.FILE_SHARE_READ | fileflags.FILE_SHARE_WRITE
//	CreateDisposition: fileflags.FILE_OPEN
//	CreateOptions:     fileflags.FILE_NON_DIRECTORY_FILE
package fileflags

// Generic and standard access rights of the access mask.
//
// [MS-DTYP] 2.4.3 ACCESS_MASK:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/7a53f60e-e730-4dfe-bbe9-b21b62eb790b
const (
	// Generic rights (bits 28-31), mapped by the resource manager to object-
	// specific rights.
	GENERIC_ALL     uint32 = 0x10000000
	GENERIC_EXECUTE uint32 = 0x20000000
	GENERIC_WRITE   uint32 = 0x40000000
	GENERIC_READ    uint32 = 0x80000000

	// Standard rights (bits 16-20), common to all object types.
	DELETE       uint32 = 0x00010000
	READ_CONTROL uint32 = 0x00020000
	WRITE_DAC    uint32 = 0x00040000
	WRITE_OWNER  uint32 = 0x00080000
	SYNCHRONIZE  uint32 = 0x00100000
)

// File, pipe, and printer specific access rights, and the directory-context
// aliases that share the same bit values (e.g. FILE_LIST_DIRECTORY ==
// FILE_READ_DATA). Use these in the DesiredAccess field.
//
// [MS-SMB2] 2.2.13.1.1 File_Pipe_Printer_Access_Mask:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/77b36d0f-6016-458a-a7a0-0f4a72ae1534
//
// [MS-SMB2] 2.2.13.1.2 Directory_Access_Mask:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/0a5934b1-80f1-4da0-b1bf-5e021c309b71
const (
	FILE_READ_DATA        uint32 = 0x00000001
	FILE_LIST_DIRECTORY   uint32 = 0x00000001
	FILE_WRITE_DATA       uint32 = 0x00000002
	FILE_ADD_FILE         uint32 = 0x00000002
	FILE_APPEND_DATA      uint32 = 0x00000004
	FILE_ADD_SUBDIRECTORY uint32 = 0x00000004
	FILE_READ_EA          uint32 = 0x00000008
	FILE_WRITE_EA         uint32 = 0x00000010
	FILE_EXECUTE          uint32 = 0x00000020
	FILE_TRAVERSE         uint32 = 0x00000020
	FILE_DELETE_CHILD     uint32 = 0x00000040
	FILE_READ_ATTRIBUTES  uint32 = 0x00000080
	FILE_WRITE_ATTRIBUTES uint32 = 0x00000100
)

// ShareAccess — the sharing mode granted to other openers.
//
// [MS-SMB2] 2.2.13 SMB2 CREATE Request:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/e8fb45c1-a03d-44ca-b7ae-47385cfd7997
const (
	FILE_SHARE_NONE   uint32 = 0x00000000
	FILE_SHARE_READ   uint32 = 0x00000001
	FILE_SHARE_WRITE  uint32 = 0x00000002
	FILE_SHARE_DELETE uint32 = 0x00000004
)

// CreateDisposition — the action taken depending on whether the file exists.
//
// [MS-SMB2] 2.2.13 SMB2 CREATE Request:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/e8fb45c1-a03d-44ca-b7ae-47385cfd7997
const (
	FILE_SUPERSEDE    uint32 = 0x00000000
	FILE_OPEN         uint32 = 0x00000001
	FILE_CREATE       uint32 = 0x00000002
	FILE_OPEN_IF      uint32 = 0x00000003
	FILE_OVERWRITE    uint32 = 0x00000004
	FILE_OVERWRITE_IF uint32 = 0x00000005
)

// CreateOptions — options applied when opening.
//
// [MS-SMB2] 2.2.13 SMB2 CREATE Request:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/e8fb45c1-a03d-44ca-b7ae-47385cfd7997
const (
	FILE_DIRECTORY_FILE          uint32 = 0x00000001
	FILE_WRITE_THROUGH           uint32 = 0x00000002
	FILE_SEQUENTIAL_ONLY         uint32 = 0x00000004
	FILE_NO_INTERMEDIATE_BUFFER  uint32 = 0x00000008
	FILE_SYNCHRONOUS_IO_ALERT    uint32 = 0x00000010
	FILE_SYNCHRONOUS_IO_NONALERT uint32 = 0x00000020
	FILE_NON_DIRECTORY_FILE      uint32 = 0x00000040
	FILE_RANDOM_ACCESS           uint32 = 0x00000800
	FILE_DELETE_ON_CLOSE         uint32 = 0x00001000
	FILE_OPEN_BY_FILE_ID         uint32 = 0x00002000
	FILE_OPEN_FOR_BACKUP_INTENT  uint32 = 0x00004000
)

// FileAttributes — file attribute flags.
//
// [MS-FSCC] 2.6 File Attributes:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/ca28ec38-f155-4768-81d6-4bfeb8586fc9
const (
	FILE_ATTRIBUTE_READONLY  uint32 = 0x00000001
	FILE_ATTRIBUTE_HIDDEN    uint32 = 0x00000002
	FILE_ATTRIBUTE_SYSTEM    uint32 = 0x00000004
	FILE_ATTRIBUTE_DIRECTORY uint32 = 0x00000010
	FILE_ATTRIBUTE_ARCHIVE   uint32 = 0x00000020
	FILE_ATTRIBUTE_NORMAL    uint32 = 0x00000080
)
