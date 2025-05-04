package data_structures

import "github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"

// SERVER_INFO_101
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/39c502dd-022b-4a68-9367-89fd76a23bc3
type SERVER_INFO_101 struct {
	// sv101_platform_id: Specifies the information level to use for platform-specific information.
	Sv101PlatformId data_types.DWORD
	// sv101_name: A pointer to a null-terminated Unicode UTF-16 Internet host name or NetBIOS host name of a server.
	Sv101Name data_types.STRING
	// sv101_version_major: Specifies the major release version number of the operating system. The server MUST set this
	// field to an implementation-specific major release version number that corresponds to the host operating system as
	// specified in the following table.
	Sv101VersionMajor data_types.DWORD
	// sv101_version_minor: Specifies the minor release version number of the operating system. The server MUST set this
	// field to an implementation-specific minor release version number that corresponds to the host operating system as
	// specified in the following table.
	Sv101VersionMinor data_types.DWORD
	// sv101_version_type: The sv101_version_type field specifies the SV_TYPE flags, which indicate the software services
	// that are available (but not necessarily running) on the server. This member MUST be a combination of one or more of
	// the following values.
	Sv101VersionType data_types.DWORD
	// sv101_comment: A pointer to a null-terminated Unicode UTF-16 string that specifies a comment that describes the server.
	Sv101Comment data_types.STRING
}
