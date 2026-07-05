package mstsgu

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TSG_PACKET_TYPE_UNION is the [switch_type(unsigned long)] union that carries the body
// of a TSG_PACKET ([MS-TSGU] 2.2.9.2.1.9). Tag carries the discriminant (packetId)
// inline, followed by the selected arm ([C706] 14.3.8). Each arm is a [unique] pointer
// to its packet structure, as declared in the IDL.
type TSG_PACKET_TYPE_UNION struct {
	Tag ndr.DWORD `ndr:"switch"`
	// case TSG_PACKET_TYPE_HEADER (0x00004844)
	PacketHeader *TSG_PACKET_HEADER `ndr:"case=0x00004844,unique"`
	// case TSG_PACKET_TYPE_VERSIONCAPS (0x00005643)
	PacketVersionCaps *TSG_PACKET_VERSIONCAPS `ndr:"case=0x00005643,unique"`
	// case TSG_PACKET_TYPE_QUARCONFIGREQUEST (0x00005143)
	PacketQuarConfigRequest *TSG_PACKET_QUARCONFIGREQUEST `ndr:"case=0x00005143,unique"`
	// case TSG_PACKET_TYPE_QUARREQUEST (0x00005152)
	PacketQuarRequest *TSG_PACKET_QUARREQUEST `ndr:"case=0x00005152,unique"`
	// case TSG_PACKET_TYPE_RESPONSE (0x00005052)
	PacketResponse *TSG_PACKET_RESPONSE `ndr:"case=0x00005052,unique"`
	// case TSG_PACKET_TYPE_QUARENC_RESPONSE (0x00004552)
	PacketQuarEncResponse *TSG_PACKET_QUARENC_RESPONSE `ndr:"case=0x00004552,unique"`
	// case TSG_PACKET_TYPE_CAPS_RESPONSE (0x00004350)
	PacketCapsResponse *TSG_PACKET_CAPS_RESPONSE `ndr:"case=0x00004350,unique"`
	// case TSG_PACKET_TYPE_MSGREQUEST_PACKET (0x00004752)
	PacketMsgRequest *TSG_PACKET_MSG_REQUEST `ndr:"case=0x00004752,unique"`
	// case TSG_PACKET_TYPE_MESSAGE_PACKET (0x00004750)
	PacketMsgResponse *TSG_PACKET_MSG_RESPONSE `ndr:"case=0x00004750,unique"`
	// case TSG_PACKET_TYPE_AUTH (0x00004054)
	PacketAuth *TSG_PACKET_AUTH `ndr:"case=0x00004054,unique"`
	// case TSG_PACKET_TYPE_REAUTH (0x00005250)
	PacketReauth *TSG_PACKET_REAUTH `ndr:"case=0x00005250,unique"`
}
