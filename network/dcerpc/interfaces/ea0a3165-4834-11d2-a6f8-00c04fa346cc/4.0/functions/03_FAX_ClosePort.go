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

// fAX_ClosePortRequest carries the [in] parameters of FAX_ClosePort.
type fAX_ClosePortRequest struct {
	FaxPortHandle msfax.PRPC_FAX_PORT_HANDLE
}

func (*fAX_ClosePortRequest) Opnum() uint16 { return fax.OpnumFAX_ClosePort }

// fAX_ClosePortResponse carries the [out] parameters and return value of FAX_ClosePort.
type fAX_ClosePortResponse struct {
	FaxPortHandle msfax.PRPC_FAX_PORT_HANDLE
	Status        ndr.DWORD `ndr:"retval"`
}

// FAX_ClosePort calls FAX_ClosePort (opnum 3) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_ClosePort(rpc ndr.Invoker, faxPortHandle msfax.PRPC_FAX_PORT_HANDLE) (FaxPortHandle msfax.PRPC_FAX_PORT_HANDLE, err error) {
	req := &fAX_ClosePortRequest{
		FaxPortHandle: faxPortHandle,
	}
	var resp fAX_ClosePortResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_ClosePort: %w", err)
		return
	}
	FaxPortHandle = resp.FaxPortHandle
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_ClosePort failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
