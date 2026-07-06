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

// fAX_OpenPortRequest carries the [in] parameters of FAX_OpenPort.
type fAX_OpenPortRequest struct {
	DeviceId ndr.DWORD
	Flags    ndr.DWORD
}

func (*fAX_OpenPortRequest) Opnum() uint16 { return fax.OpnumFAX_OpenPort }

// fAX_OpenPortResponse carries the [out] parameters and return value of FAX_OpenPort.
type fAX_OpenPortResponse struct {
	FaxPortHandle msfax.PRPC_FAX_PORT_HANDLE
	Status        ndr.DWORD `ndr:"retval"`
}

// FAX_OpenPort calls FAX_OpenPort (opnum 2) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_OpenPort(rpc ndr.Invoker, deviceId ndr.DWORD, flags ndr.DWORD) (FaxPortHandle msfax.PRPC_FAX_PORT_HANDLE, err error) {
	req := &fAX_OpenPortRequest{
		DeviceId: deviceId,
		Flags:    flags,
	}
	var resp fAX_OpenPortResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_OpenPort: %w", err)
		return
	}
	FaxPortHandle = resp.FaxPortHandle
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_OpenPort failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
