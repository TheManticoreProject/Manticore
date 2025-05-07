package types

// SMB_EXT_FILE_ATTR
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/6008aa8f-d2d8-4366-b775-b81aece05bb1
// A 32-bit field containing encoded file attribute values and file access behavior flag values. The attribute and
// flag value names are for reference purposes only. If ATTR_NORMAL (see following) is set as the requested attribute
// value, it MUST be the only attribute value set. Including any other attribute value causes the ATTR_NORMAL value to
// be ignored. Any combination of the flag values (see following) is acceptable.
type SMB_EXT_FILE_ATTR DWORD

const (
	// ATTR_READONLY - The file is read only. Applications can read the file but cannot write to it or delete it.
	ATTR_READONLY SMB_EXT_FILE_ATTR = 0x00000001

	// ATTR_HIDDEN - The file is hidden. It is not to be included in an ordinary directory listing.
	ATTR_HIDDEN SMB_EXT_FILE_ATTR = 0x00000002

	// ATTR_SYSTEM - The file is part of or is used exclusively by the operating system.
	ATTR_SYSTEM SMB_EXT_FILE_ATTR = 0x00000004

	// ATTR_DIRECTORY - The file is a directory.
	ATTR_DIRECTORY SMB_EXT_FILE_ATTR = 0x00000010

	// ATTR_ARCHIVE - The file has not been archived since it was last modified.
	ATTR_ARCHIVE SMB_EXT_FILE_ATTR = 0x00000020

	// ATTR_NORMAL - The file has no other attributes set. This attribute is valid only if used alone.
	ATTR_NORMAL SMB_EXT_FILE_ATTR = 0x00000080

	// ATTR_TEMPORARY - The file is temporary. This is a hint to the cache manager that it does not need to flush the file to backing storage.
	ATTR_TEMPORARY SMB_EXT_FILE_ATTR = 0x00000100

	// ATTR_COMPRESSED - The file or directory is compressed. For a file, this means that all of the data in the file is compressed. For a directory, this means that compression is the default for newly created files and subdirectories.
	ATTR_COMPRESSED SMB_EXT_FILE_ATTR = 0x00000800

	// POSIX_SEMANTICS - Indicates that the file is to be accessed according to POSIX rules. This includes allowing multiple files with names differing only in case, for file systems that support such naming.
	POSIX_SEMANTICS SMB_EXT_FILE_ATTR = 0x01000000

	// BACKUP_SEMANTICS - Indicates that the file is being opened or created for a backup or restore operation. The server SHOULD allow the client to override normal file security checks, provided it has the necessary permission to do so.
	BACKUP_SEMANTICS SMB_EXT_FILE_ATTR = 0x02000000

	// DELETE_ON_CLOSE - Requests that the server delete the file immediately after all of its handles have been closed.
	DELETE_ON_CLOSE SMB_EXT_FILE_ATTR = 0x04000000

	// SEQUENTIAL_SCAN - Indicates that the file is to be accessed sequentially from beginning to end.
	SEQUENTIAL_SCAN SMB_EXT_FILE_ATTR = 0x08000000

	// RANDOM_ACCESS - Indicates that the application is designed to access the file randomly. The server can use this flag to optimize file caching.
	RANDOM_ACCESS SMB_EXT_FILE_ATTR = 0x10000000

	// NO_BUFFERING - Requests that the server open the file with no intermediate buffering or caching; the server might not honor the request. The application MUST meet certain requirements when working with files opened with FILE_FLAG_NO_BUFFERING. File access MUST begin at offsets within the file that are integer multiples of the volume's sector size and MUST be for numbers of bytes that are integer multiples of the volume's sector size. For example, if the sector size is 512 bytes, an application can request reads and writes of 512, 1024, or 2048 bytes, but not of 335, 981, or 7171 bytes.
	NO_BUFFERING SMB_EXT_FILE_ATTR = 0x20000000

	// WRITE_THROUGH - Instructs the operating system to write through any intermediate cache and go directly to the file. The operating system can still cache write operations, but cannot lazily flush them.
	WRITE_THROUGH SMB_EXT_FILE_ATTR = 0x80000000
)
