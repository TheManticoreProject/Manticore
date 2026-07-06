package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_ReAssignMessageRequest carries the [in] parameters of FAX_ReAssignMessage.
type fAX_ReAssignMessageRequest struct {
	DwlMessageId  uint64
	PReAssignInfo msfax.FAX_REASSIGN_INFO
}

func (*fAX_ReAssignMessageRequest) Opnum() uint16 { return fax.OpnumFAX_ReAssignMessage }

// fAX_ReAssignMessageResponse carries the [out] parameters and return value of FAX_ReAssignMessage.
type fAX_ReAssignMessageResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_ReAssignMessage calls FAX_ReAssignMessage (opnum 101) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_ReAssignMessage(rpc ndr.Invoker, dwlMessageId uint64, pReAssignInfo msfax.FAX_REASSIGN_INFO) (err error) {
	req := &fAX_ReAssignMessageRequest{
		DwlMessageId:  dwlMessageId,
		PReAssignInfo: pReAssignInfo,
	}
	var resp fAX_ReAssignMessageResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_ReAssignMessage: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_ReAssignMessage failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
