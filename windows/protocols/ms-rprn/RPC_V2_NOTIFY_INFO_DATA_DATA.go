package msrprn

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// RPC_V2_NOTIFY_INFO_DATA_DATA is the [switch_is(Reserved & 0xFFFF)] union embedded in
// RPC_V2_NOTIFY_INFO_DATA ([MS-RPRN] 2.2.1.13.4). The discriminant is the low 16 bits of
// Reserved, transmitted as a DWORD; the case values are the [MS-RPRN] TABLE_* #define
// constants (TABLE_DWORD=0x1 … TABLE_SECURITYDESCRIPTOR=0x5). The ndr codec parses case
// labels as integers, so they are numeric here rather than the symbolic names.
type RPC_V2_NOTIFY_INFO_DATA_DATA struct {
	Tag                ndr.DWORD            `ndr:"switch"`
	DwData             ndr.DWORD            `ndr:"case=1"` // TABLE_DWORD
	String             STRING_CONTAINER     `ndr:"case=2"` // TABLE_STRING
	DevMode            DEVMODE_CONTAINER    `ndr:"case=3"` // TABLE_DEVMODE
	SystemTime         SYSTEMTIME_CONTAINER `ndr:"case=4"` // TABLE_TIME
	SecurityDescriptor SECURITY_CONTAINER   `ndr:"case=5"` // TABLE_SECURITYDESCRIPTOR
}
