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

// fAX_StartServerNotificationEx2Request carries the [in] parameters of FAX_StartServerNotificationEx2.
type fAX_StartServerNotificationEx2Request struct {
	LpcwstrAccountName   *ndr.WSTR `ndr:"unique"`
	LpcwstrMachineName   ndr.WSTR
	LpcwstrEndPoint      ndr.WSTR
	Context              uint64
	LpcwstrProtseqString ndr.WSTR
	DwEventTypes         ndr.DWORD
	Level                ndr.DWORD
}

func (*fAX_StartServerNotificationEx2Request) Opnum() uint16 {
	return fax.OpnumFAX_StartServerNotificationEx2
}

// fAX_StartServerNotificationEx2Response carries the [out] parameters and return value of FAX_StartServerNotificationEx2.
type fAX_StartServerNotificationEx2Response struct {
	LpHandle msfax.PRPC_FAX_EVENT_EX_HANDLE
	Status   ndr.DWORD `ndr:"retval"`
}

// FAX_StartServerNotificationEx2 calls FAX_StartServerNotificationEx2 (opnum 91) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_StartServerNotificationEx2(rpc ndr.Invoker, lpcwstrAccountName *ndr.WSTR, lpcwstrMachineName ndr.WSTR, lpcwstrEndPoint ndr.WSTR, context uint64, lpcwstrProtseqString ndr.WSTR, dwEventTypes ndr.DWORD, level ndr.DWORD) (LpHandle msfax.PRPC_FAX_EVENT_EX_HANDLE, err error) {
	req := &fAX_StartServerNotificationEx2Request{
		LpcwstrAccountName:   lpcwstrAccountName,
		LpcwstrMachineName:   lpcwstrMachineName,
		LpcwstrEndPoint:      lpcwstrEndPoint,
		Context:              context,
		LpcwstrProtseqString: lpcwstrProtseqString,
		DwEventTypes:         dwEventTypes,
		Level:                level,
	}
	var resp fAX_StartServerNotificationEx2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_StartServerNotificationEx2: %w", err)
		return
	}
	LpHandle = resp.LpHandle
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_StartServerNotificationEx2 failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
