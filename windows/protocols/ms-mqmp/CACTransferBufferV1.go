package msmqmp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// CACTransferBufferV1 carries the parameters and message properties of a Send, Receive, or
// CreateCursor operation ([MS-MQMP] 2.2.5). It is a large, heavily optional structure: a
// uTransferType discriminant selects one of three operation-specific arms, and the trailing
// members carry every message property as an independent [unique] pointer.
//
// NOT YET WIRE-READY. Faithful NDR marshalling of this structure requires pointer
// double-indirection that the declarative codec (network/dcerpc/ndr) does not yet support:
// the message-property buffers are IDL "[size_is(,n)] TYPE**" — a unique pointer to a
// unique pointer to a conformant[-varying] array — and the id fields are "TYPE**" (a unique
// pointer to a unique pointer to a single value). The codec models only a single level of
// pointer-to-array / pointer-to-value, so the fields below are declared with the closest
// single-indirection form for documentation and Go-level use, but they are known to be one
// referent id short of the wire and are NOT round-trip tested. The four methods that carry
// a CACTransferBuffer (R_QMCreateRemoteCursor, and qmcomm2's QMSendMessageInternalEx /
// rpc_ACSendMessageEx / rpc_ACReceiveMessageEx) are deferred until the codec gains
// double-indirection support and the layout can be validated against a live MSMQ server.
type CACTransferBufferV1 struct {
	UTransferType ndr.DWORD
	Value         CACTransferBufferV1Union

	PClass          *uint16          `ndr:"unique"`
	PpMessageID     *msmqmq.OBJECTID `ndr:"unique"`
	PpCorrelationID []byte           `ndr:"unique,varying,size_is=20,length_is=20"`
	PSentTime       *ndr.DWORD       `ndr:"unique"`
	PArrivedTime    *ndr.DWORD       `ndr:"unique"`
	PPriority       *uint8           `ndr:"unique"`
	PDelivery       *uint8           `ndr:"unique"`
	PAcknowledge    *uint8           `ndr:"unique"`
	PAuditing       *uint8           `ndr:"unique"`
	PApplicationTag *ndr.DWORD       `ndr:"unique"`

	PpBody                   []byte `ndr:"unique,varying,size_is=UlAllocBodyBufferInBytes,length_is=UlBodyBufferSizeInBytes"`
	UlBodyBufferSizeInBytes  ndr.DWORD
	UlAllocBodyBufferInBytes ndr.DWORD
	PBodySize                *ndr.DWORD `ndr:"unique"`

	PpTitle                    []uint16 `ndr:"unique,varying,size_is=UlTitleBufferSizeInWCHARs,length_is=UlTitleBufferSizeInWCHARs"`
	UlTitleBufferSizeInWCHARs  ndr.DWORD
	PulTitleBufferSizeInWCHARs *ndr.DWORD `ndr:"unique"`

	UlAbsoluteTimeToQueue  ndr.DWORD
	PulRelativeTimeToQueue *ndr.DWORD `ndr:"unique"`
	UlRelativeTimeToLive   ndr.DWORD
	PulRelativeTimeToLive  *ndr.DWORD `ndr:"unique"`
	PTrace                 *uint8     `ndr:"unique"`

	PulSenderIDType    *ndr.DWORD `ndr:"unique"`
	PpSenderID         []byte     `ndr:"unique,size_is=USenderIDLen"`
	PulSenderIDLenProp *ndr.DWORD `ndr:"unique"`
	PulPrivLevel       *ndr.DWORD `ndr:"unique"`
	UlAuthLevel        ndr.DWORD
	PAuthenticated     *uint8     `ndr:"unique"`
	PulHashAlg         *ndr.DWORD `ndr:"unique"`
	PulEncryptAlg      *ndr.DWORD `ndr:"unique"`

	PpSenderCert         []byte `ndr:"unique,size_is=UlSenderCertLen"`
	UlSenderCertLen      ndr.DWORD
	PulSenderCertLenProp *ndr.DWORD `ndr:"unique"`

	PpwcsProvName          []uint16 `ndr:"unique,size_is=UlProvNameLen"`
	UlProvNameLen          ndr.DWORD
	PulAuthProvNameLenProp *ndr.DWORD `ndr:"unique"`
	PulProvType            *ndr.DWORD `ndr:"unique"`
	FDefaultProvider       int32

	PpSymmKeys          []byte `ndr:"unique,size_is=UlSymmKeysSize"`
	UlSymmKeysSize      ndr.DWORD
	PulSymmKeysSizeProp *ndr.DWORD `ndr:"unique"`
	BEncrypted          uint8
	BAuthenticated      uint8
	USenderIDLen        uint16

	PpSignature          []byte `ndr:"unique,size_is=UlSignatureSize"`
	UlSignatureSize      ndr.DWORD
	PulSignatureSizeProp *ndr.DWORD      `ndr:"unique"`
	PpSrcQMID            *msdtyp.GUID    `ndr:"unique"`
	PUow                 *msmqmq.XACTUOW `ndr:"unique"`

	PpMsgExtension              []byte `ndr:"unique,varying,size_is=UlMsgExtensionBufferInBytes,length_is=UlMsgExtensionBufferInBytes"`
	UlMsgExtensionBufferInBytes ndr.DWORD
	PMsgExtensionSize           *ndr.DWORD   `ndr:"unique"`
	PpConnectorType             *msdtyp.GUID `ndr:"unique"`
	PulBodyType                 *ndr.DWORD   `ndr:"unique"`
	PulVersion                  *ndr.DWORD   `ndr:"unique"`
}

// CACTransferBufferV1Union is the [switch_is(uTransferType)] union of CACTransferBufferV1
// ([MS-MQMP] 2.2.5). CACTB_SEND (0) selects the Send arm, CACTB_RECEIVE (1) the Receive
// arm, and CACTB_CREATECURSOR (2) a CACCreateRemoteCursor.
type CACTransferBufferV1Union struct {
	UTransferType ndr.DWORD `ndr:"switch"`

	Send         CACTransferBufferV1Send    `ndr:"case=0"`
	Receive      CACTransferBufferV1Receive `ndr:"case=1"`
	CreateCursor CACCreateRemoteCursor      `ndr:"case=2"`
}

// CACTransferBufferV1Send is the CACTB_SEND arm ([MS-MQMP] 2.2.5): the admin and response
// queue format names for a Send operation.
type CACTransferBufferV1Send struct {
	PAdminQueueFormat    *msmqmq.QUEUE_FORMAT `ndr:"unique"`
	PResponseQueueFormat *msmqmq.QUEUE_FORMAT `ndr:"unique"`
}

// CACTransferBufferV1Receive is the CACTB_RECEIVE arm ([MS-MQMP] 2.2.5): the parameters of a
// Receive operation. The four format-name buffers are IDL "[size_is(,n)] WCHAR**" (see the
// double-indirection note on CACTransferBufferV1) and are modelled here with the closest
// single-indirection form.
type CACTransferBufferV1Receive struct {
	RequestTimeout               ndr.DWORD
	Action                       ndr.DWORD
	Asynchronous                 ndr.DWORD
	Cursor                       ndr.DWORD
	UlResponseFormatNameLen      ndr.DWORD
	PpResponseFormatName         []uint16   `ndr:"unique,size_is=UlResponseFormatNameLen"`
	PulResponseFormatNameLenProp *ndr.DWORD `ndr:"unique"`
	UlAdminFormatNameLen         ndr.DWORD
	PpAdminFormatName            []uint16   `ndr:"unique,size_is=UlAdminFormatNameLen"`
	PulAdminFormatNameLenProp    *ndr.DWORD `ndr:"unique"`
	UlDestFormatNameLen          ndr.DWORD
	PpDestFormatName             []uint16   `ndr:"unique,size_is=UlDestFormatNameLen"`
	PulDestFormatNameLenProp     *ndr.DWORD `ndr:"unique"`
	UlOrderingFormatNameLen      ndr.DWORD
	PpOrderingFormatName         []uint16   `ndr:"unique,size_is=UlOrderingFormatNameLen"`
	PulOrderingFormatNameLenProp *ndr.DWORD `ndr:"unique"`
}
