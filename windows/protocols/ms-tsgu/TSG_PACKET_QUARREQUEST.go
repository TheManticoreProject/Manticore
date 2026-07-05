package mstsgu

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TSG_PACKET_QUARREQUEST is the health/quarantine request packet ([MS-TSGU]
// 2.2.9.2.1.4). MachineName is a [string] (conformant-varying) wide-char array sized by
// NameLength; Data is a [unique] conformant byte blob sized by DataLen.
type TSG_PACKET_QUARREQUEST struct {
	Flags       ndr.DWORD
	MachineName []uint16 `ndr:"unique,varying,size_is=NameLength"`
	NameLength  ndr.DWORD
	Data        []uint8 `ndr:"unique,size_is=DataLen"`
	DataLen     ndr.DWORD
}
