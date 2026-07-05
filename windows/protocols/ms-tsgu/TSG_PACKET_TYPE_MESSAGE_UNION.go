package mstsgu

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TSG_PACKET_TYPE_MESSAGE_UNION is the [switch_type(unsigned long)] union of
// administrative-message packets ([MS-TSGU] 2.2.9.2.1.8). Tag carries the discriminant
// (msgType) inline, followed by the selected arm ([C706] 14.3.8). Each arm is a
// [unique] pointer to its message structure, as declared in the IDL.
type TSG_PACKET_TYPE_MESSAGE_UNION struct {
	Tag ndr.DWORD `ndr:"switch"`
	// case TSG_ASYNC_MESSAGE_CONSENT_MESSAGE (0x00000001)
	ConsentMessage *TSG_PACKET_STRING_MESSAGE `ndr:"case=0x00000001,unique"`
	// case TSG_ASYNC_MESSAGE_SERVICE_MESSAGE (0x00000002)
	ServiceMessage *TSG_PACKET_STRING_MESSAGE `ndr:"case=0x00000002,unique"`
	// case TSG_ASYNC_MESSAGE_REAUTH (0x00000003)
	ReauthMessage *TSG_PACKET_REAUTH_MESSAGE `ndr:"case=0x00000003,unique"`
}
