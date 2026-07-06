package functions

// IDL source: [MS-MQDS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqds/7907bc25-e4e6-40ef-b990-9172d1808e94
// A fetched copy is kept at ms-mqds.idl in the interface directory.

import (
	"fmt"

	dscomm2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/708cca10-9569-11d1-b2a5-0060977d8118/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
)

// s_DSBeginDeleteNotificationRequest carries the [in] parameters of S_DSBeginDeleteNotification.
type s_DSBeginDeleteNotificationRequest struct {
	PwcsPathName ndr.WSTR
	PhServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE
}

func (*s_DSBeginDeleteNotificationRequest) Opnum() uint16 {
	return dscomm2.OpnumS_DSBeginDeleteNotification
}

// s_DSBeginDeleteNotificationResponse carries the [out] parameters and return value of S_DSBeginDeleteNotification.
type s_DSBeginDeleteNotificationResponse struct {
	PHandle msmqds.PPCONTEXT_HANDLE_DELETE_TYPE
	Status  ndr.DWORD `ndr:"retval"`
}

// S_DSBeginDeleteNotification calls S_DSBeginDeleteNotification (opnum 3) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSBeginDeleteNotification(rpc ndr.Invoker, pwcsPathName ndr.WSTR, phServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE) (PHandle msmqds.PPCONTEXT_HANDLE_DELETE_TYPE, err error) {
	req := &s_DSBeginDeleteNotificationRequest{
		PwcsPathName: pwcsPathName,
		PhServerAuth: phServerAuth,
	}
	var resp s_DSBeginDeleteNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSBeginDeleteNotification: %w", err)
		return
	}
	PHandle = resp.PHandle
	if uint32(resp.Status) != dscomm2.StatusSuccess {
		err = fmt.Errorf("S_DSBeginDeleteNotification failed: %s", dscomm2.StatusString(uint32(resp.Status)))
	}
	return
}
