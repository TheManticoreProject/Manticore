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

// fAX_StartServerNotificationExRequest carries the [in] parameters of FAX_StartServerNotificationEx.
type fAX_StartServerNotificationExRequest struct {
	LpcwstrMachineName ndr.WSTR
	LpcwstrEndPoint    ndr.WSTR
	Context            uint64
	LpcwstrProtSeq     ndr.WSTR
	BEventEx           ndr.BOOL
	DwEventTypes       ndr.DWORD
}

func (*fAX_StartServerNotificationExRequest) Opnum() uint16 {
	return fax.OpnumFAX_StartServerNotificationEx
}

// fAX_StartServerNotificationExResponse carries the [out] parameters and return value of FAX_StartServerNotificationEx.
type fAX_StartServerNotificationExResponse struct {
	LpHandle msfax.PRPC_FAX_EVENT_EX_HANDLE
	Status   ndr.DWORD `ndr:"retval"`
}

// FAX_StartServerNotificationEx calls FAX_StartServerNotificationEx (opnum 74) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_StartServerNotificationEx(rpc ndr.Invoker, lpcwstrMachineName ndr.WSTR, lpcwstrEndPoint ndr.WSTR, context uint64, lpcwstrProtSeq ndr.WSTR, bEventEx ndr.BOOL, dwEventTypes ndr.DWORD) (LpHandle msfax.PRPC_FAX_EVENT_EX_HANDLE, err error) {
	req := &fAX_StartServerNotificationExRequest{
		LpcwstrMachineName: lpcwstrMachineName,
		LpcwstrEndPoint:    lpcwstrEndPoint,
		Context:            context,
		LpcwstrProtSeq:     lpcwstrProtSeq,
		BEventEx:           bEventEx,
		DwEventTypes:       dwEventTypes,
	}
	var resp fAX_StartServerNotificationExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_StartServerNotificationEx: %w", err)
		return
	}
	LpHandle = resp.LpHandle
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_StartServerNotificationEx failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
