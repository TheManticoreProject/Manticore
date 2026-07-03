package msraiw

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// WINSINTF_RECORD_ACTION_T ([MS-RAIW] 2.2.2.1) is the record acted on by
// R_WinsRecordAction and returned by the R_WinsGetDbRecs* enumerations.
type WINSINTF_RECORD_ACTION_T struct {
	Cmd_e WINSINTF_ACT_E
	// PName is [size_is(NameLen + 1)] LPBYTE: a [unique] pointer to a conformant byte
	// buffer holding the NetBIOS name plus a trailing NUL. The size_is bound is an
	// arithmetic expression the codec cannot key on, so the maximum_count derives from
	// the slice length — callers set PName to NameLen+1 bytes and NameLen to match.
	PName      []uint8 `ndr:"unique,conformant"`
	NameLen    ndr.DWORD
	TypOfRec_e ndr.DWORD
	NoOfAdds   ndr.DWORD
	PAdd       []WINSINTF_ADD_T `ndr:"unique,size_is=NoOfAdds"`
	Add        WINSINTF_ADD_T
	VersNo     dtyp.LARGE_INTEGER
	NodeTyp    uint8
	OwnerId    ndr.DWORD
	State_e    ndr.DWORD
	FStatic    ndr.DWORD
	// TimeStamp is IDL DWORD_PTR. Under the NDR20 transfer syntax an __int3264
	// pointer-sized integer is transmitted as a 4-octet value ([MS-RPCE] 2.2.4.1),
	// so it is modeled as a 32-bit DWORD.
	TimeStamp ndr.DWORD
}
