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

// fAX_ConnectionRefCountRequest carries the [in] parameters of FAX_ConnectionRefCount.
type fAX_ConnectionRefCountRequest struct {
	Handle  msfax.PRPC_FAX_SVC_HANDLE
	Connect ndr.DWORD
}

func (*fAX_ConnectionRefCountRequest) Opnum() uint16 { return fax.OpnumFAX_ConnectionRefCount }

// fAX_ConnectionRefCountResponse carries the [out] parameters and return value of FAX_ConnectionRefCount.
type fAX_ConnectionRefCountResponse struct {
	Handle   msfax.PRPC_FAX_SVC_HANDLE
	CanShare ndr.DWORD
	Status   ndr.DWORD `ndr:"retval"`
}

// FAX_ConnectionRefCount calls FAX_ConnectionRefCount (opnum 1) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_ConnectionRefCount(rpc ndr.Invoker, handle msfax.PRPC_FAX_SVC_HANDLE, connect ndr.DWORD) (Handle msfax.PRPC_FAX_SVC_HANDLE, CanShare ndr.DWORD, err error) {
	req := &fAX_ConnectionRefCountRequest{
		Handle:  handle,
		Connect: connect,
	}
	var resp fAX_ConnectionRefCountResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_ConnectionRefCount: %w", err)
		return
	}
	Handle = resp.Handle
	CanShare = resp.CanShare
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_ConnectionRefCount failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
