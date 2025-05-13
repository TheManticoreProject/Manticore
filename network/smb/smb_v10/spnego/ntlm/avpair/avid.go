package avpair

type AvId uint16

// AV_PAIR types used in NTLM authentication
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/83f5e789-660d-4781-8491-5f8c6641f75e
const (
	// MsvAvEOL indicates that this is the last AV_PAIR in the list.
	// AvLen MUST be 0. This type of information MUST be present in the AV pair list.
	MsvAvEOL AvId = 0x0000

	// MsvAvNbComputerName is the server's NetBIOS computer name.
	// The name MUST be in Unicode, and is not null-terminated.
	// This type of information MUST be present in the AV_pair list.
	MsvAvNbComputerName AvId = 0x0001

	// MsvAvNbDomainName is the server's NetBIOS domain name.
	// The name MUST be in Unicode, and is not null-terminated.
	// This type of information MUST be present in the AV_pair list.
	MsvAvNbDomainName AvId = 0x0002

	// MsvAvDnsComputerName is the fully qualified domain name (FQDN) of the computer.
	// The name MUST be in Unicode, and is not null-terminated.
	MsvAvDnsComputerName AvId = 0x0003

	// MsvAvDnsDomainName is the FQDN of the domain.
	// The name MUST be in Unicode, and is not null-terminated.
	MsvAvDnsDomainName AvId = 0x0004

	// MsvAvDnsTreeName is the FQDN of the forest.
	// The name MUST be in Unicode, and is not null-terminated.
	MsvAvDnsTreeName AvId = 0x0005

	// MsvAvFlags is a 32-bit value indicating server or client configuration.
	// 0x00000001: Indicates to the client that the account authentication is constrained.
	// 0x00000002: Indicates that the client is providing message integrity in the MIC field in the AUTHENTICATE_MESSAGE.
	// 0x00000004: Indicates that the client is providing a target SPN generated from an untrusted source.
	MsvAvFlags AvId = 0x0006

	// MsvAvTimestamp is a FILETIME structure in little-endian byte order that contains the server local time.
	// This structure is always sent in the CHALLENGE_MESSAGE.
	MsvAvTimestamp AvId = 0x0007

	// MsvAvSingleHost is a Single_Host_Data structure.
	// The Value field contains a platform-specific blob, as well as a MachineID created at computer startup to identify the calling machine.
	MsvAvSingleHost AvId = 0x0008

	// MsvAvTargetName is the SPN of the target server.
	// The name MUST be in Unicode and is not null-terminated.
	MsvAvTargetName AvId = 0x0009

	// MsvAvChannelBindings is a channel bindings hash.
	// The Value field contains an MD5 hash of a gss_channel_bindings_struct.
	// An all-zero value of the hash is used to indicate absence of channel bindings.
	MsvAvChannelBindings AvId = 0x000A
)
