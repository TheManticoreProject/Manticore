package mstsgu

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TSG_PACKET_QUARENC_RESPONSE is the quarantine-encapsulated response ([MS-TSGU]
// 2.2.9.2.1.5). CertChainData is a [string] (conformant-varying) wide-char array sized
// by CertChainLen; VersionCaps is a [unique] pointer to the server's version caps.
//
// Nonce is dtyp.GUID, the 16-octet NDR GUID (Data1/2/3 + Data4[8]). windows/guid.GUID
// must NOT be used on the wire: its uint64 tail over-aligns the struct and marshals to
// 24 octets under NDR. Convert with dtyp.NewGUID / (dtyp.GUID).GUID().
type TSG_PACKET_QUARENC_RESPONSE struct {
	Flags         ndr.DWORD
	CertChainLen  ndr.DWORD
	CertChainData []uint16 `ndr:"unique,varying,size_is=CertChainLen"`
	Nonce         dtyp.GUID
	VersionCaps   *TSG_PACKET_VERSIONCAPS `ndr:"unique"`
}
