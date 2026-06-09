// Package infoclass holds the [MS-FSCC] FILE_INFORMATION_CLASS and
// FILE_FS_*_INFORMATION class numbers used with SMB2 QUERY_INFO / SET_INFO. They
// live in their own package so the class identifiers can keep their spec names
// without colliding with the like-named structures in package filesystem.
package infoclass

// FileInformationClass identifies a FILE_INFORMATION_CLASS used with the
// SMB2_0_INFO_FILE info type.
//
// [MS-FSCC] 2.4 File Information Classes:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/4718fc40-e539-4014-8e33-b675af74e3e1
type FileInformationClass uint8

const (
	FileDirectoryInformation       FileInformationClass = 1
	FileFullDirectoryInformation   FileInformationClass = 2
	FileBothDirectoryInformation   FileInformationClass = 3
	FileBasicInformation           FileInformationClass = 4
	FileStandardInformation        FileInformationClass = 5
	FileInternalInformation        FileInformationClass = 6
	FileEaInformation              FileInformationClass = 7
	FileAccessInformation          FileInformationClass = 8
	FileNameInformation            FileInformationClass = 9
	FileRenameInformation          FileInformationClass = 10
	FileLinkInformation            FileInformationClass = 11
	FileNamesInformation           FileInformationClass = 12
	FileDispositionInformation     FileInformationClass = 13
	FilePositionInformation        FileInformationClass = 14
	FileAllInformation             FileInformationClass = 18
	FileAllocationInformation      FileInformationClass = 19
	FileEndOfFileInformation       FileInformationClass = 20
	FileAlternateNameInformation   FileInformationClass = 21
	FileIdBothDirectoryInformation FileInformationClass = 37
	FileIdFullDirectoryInformation FileInformationClass = 38
)

// FsInformationClass identifies a FILE_FS_*_INFORMATION class used with the
// SMB2_0_INFO_FILESYSTEM info type.
//
// [MS-FSCC] 2.5 File System Information Classes:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/ee12042a-9352-46e3-9f67-c094b75fe6c3
type FsInformationClass uint8

const (
	FileFsVolumeInformation    FsInformationClass = 1
	FileFsLabelInformation     FsInformationClass = 2
	FileFsSizeInformation      FsInformationClass = 3
	FileFsDeviceInformation    FsInformationClass = 4
	FileFsAttributeInformation FsInformationClass = 5
	FileFsControlInformation   FsInformationClass = 6
	FileFsFullSizeInformation  FsInformationClass = 7
	FileFsObjectIdInformation  FsInformationClass = 8
)
