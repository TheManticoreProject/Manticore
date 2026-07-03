package msmqqp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// REMOTEREADDESC ([MS-MQQP] 2.2.2.1) encapsulates the parameters of the RemoteQMStartReceive
// family of calls. dwSize (valid range 0..0x00420000) sizes the unique lpBuffer, which carries
// the UserMessage Packet; the client sends dwSize=0 and lpBuffer=NULL and the server fills them
// in. eAckNack is a reserved field. lpBuffer is a unique conformant-varying byte array whose
// conformance and length both come from dwSize.
type REMOTEREADDESC struct {
	HRemoteQueue ndr.DWORD
	HCursor      ndr.DWORD
	UlAction     ndr.DWORD
	UlTimeout    ndr.DWORD
	DwSize       ndr.DWORD
	DwQueue      ndr.DWORD
	DwRequestID  ndr.DWORD
	Reserved     ndr.DWORD
	DwArriveTime ndr.DWORD
	EAckNack     REMOTEREADACK `ndr:"enum"`
	LpBuffer     []uint8       `ndr:"unique,size_is=DwSize,varying,length_is=DwSize"`
}
