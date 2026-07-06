package functions

// IDL source: [MS-CMPO] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmpo/3a4677f0-9aef-41f9-9bca-f9c2469cefa6
// A fetched copy is kept at ms-cmpo.idl in the interface directory.

import (
	"fmt"

	IXnRemote "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/906b0ce0-c70b-1067-b317-00dd010662da/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmpo "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmpo"
)

// sendReceiveRequest carries the [in] parameters of SendReceive ([MS-CMPO] 3.4.4.4).
// rguchBoxCar is a conformant byte array of dwcbSizeOfBoxCar bytes carrying dwcMessages
// packed messages (the "box car").
type sendReceiveRequest struct {
	PhContext        mscmpo.PCONTEXT_HANDLE
	DwcMessages      ndr.DWORD
	DwcbSizeOfBoxCar ndr.DWORD
	RguchBoxCar      []uint8 `ndr:"ref,size_is=DwcbSizeOfBoxCar"`
}

func (*sendReceiveRequest) Opnum() uint16 { return IXnRemote.OpnumSendReceive }

// sendReceiveResponse carries the HRESULT return value of SendReceive.
type sendReceiveResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// SendReceive calls SendReceive (opnum 3) ([MS-CMPO] 3.4.4.4): it delivers messageCount
// messages packed into boxCar over the session bound to phContext. Per the IDL, boxCar
// must be 40..0x14000 bytes and messageCount 1..4095.
func SendReceive(rpc ndr.Invoker, phContext mscmpo.PCONTEXT_HANDLE, messageCount uint32, boxCar []byte) error {
	req := &sendReceiveRequest{
		PhContext:        phContext,
		DwcMessages:      ndr.DWORD(messageCount),
		DwcbSizeOfBoxCar: ndr.DWORD(len(boxCar)),
		RguchBoxCar:      boxCar,
	}
	var resp sendReceiveResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SendReceive: %w", err)
	}
	if uint32(resp.Status) != IXnRemote.StatusSuccess {
		return fmt.Errorf("SendReceive failed: %s", IXnRemote.StatusString(uint32(resp.Status)))
	}
	return nil
}
