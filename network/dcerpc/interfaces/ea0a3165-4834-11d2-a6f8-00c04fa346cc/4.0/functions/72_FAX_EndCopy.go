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

// fAX_EndCopyRequest carries the [in] parameters of FAX_EndCopy.
type fAX_EndCopyRequest struct {
	LphCopy msfax.PRPC_FAX_COPY_HANDLE
}

func (*fAX_EndCopyRequest) Opnum() uint16 { return fax.OpnumFAX_EndCopy }

// fAX_EndCopyResponse carries the [out] parameters and return value of FAX_EndCopy.
type fAX_EndCopyResponse struct {
	LphCopy msfax.PRPC_FAX_COPY_HANDLE
	Status  ndr.DWORD `ndr:"retval"`
}

// FAX_EndCopy calls FAX_EndCopy (opnum 72) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_EndCopy(rpc ndr.Invoker, lphCopy msfax.PRPC_FAX_COPY_HANDLE) (LphCopy msfax.PRPC_FAX_COPY_HANDLE, err error) {
	req := &fAX_EndCopyRequest{
		LphCopy: lphCopy,
	}
	var resp fAX_EndCopyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_EndCopy: %w", err)
		return
	}
	LphCopy = resp.LphCopy
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_EndCopy failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
