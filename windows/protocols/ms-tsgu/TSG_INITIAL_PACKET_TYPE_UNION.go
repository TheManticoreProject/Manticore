package mstsgu

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TSG_INITIAL_PACKET_TYPE_UNION is the [switch_type(unsigned long)] union of the
// packets a reauthentication may carry ([MS-TSGU] 2.2.9.2.1.7). Tag carries the
// discriminant (packetId) inline, followed by the selected arm ([C706] 14.3.8). Each
// arm is a [unique] pointer to its packet structure, as declared in the IDL.
type TSG_INITIAL_PACKET_TYPE_UNION struct {
	Tag ndr.DWORD `ndr:"switch"`
	// case TSG_PACKET_TYPE_VERSIONCAPS (0x00005643)
	PacketVersionCaps *TSG_PACKET_VERSIONCAPS `ndr:"case=0x00005643,unique"`
	// case TSG_PACKET_TYPE_AUTH (0x00004054)
	PacketAuth *TSG_PACKET_AUTH `ndr:"case=0x00004054,unique"`
}
