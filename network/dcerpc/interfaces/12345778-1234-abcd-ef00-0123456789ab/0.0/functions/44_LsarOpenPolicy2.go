package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarOpenPolicy2Request is the [in] parameter set of LsarOpenPolicy2: a NULL unique
// SystemName pointer, an inline LSAPR_OBJECT_ATTRIBUTES (a top-level [ref] struct), and
// the desired access mask.
type lsarOpenPolicy2Request struct {
	SystemName    *ndr.WSTR `ndr:"unique"`
	Attributes    mslsad.LSAPR_OBJECT_ATTRIBUTES
	DesiredAccess ndr.DWORD
}

func (*lsarOpenPolicy2Request) Opnum() uint16 { return lsarpc.OpnumLsarOpenPolicy2 }

// LsarOpenPolicy2 calls LsarOpenPolicy2 (opnum 44) and returns a policy handle.
func LsarOpenPolicy2(rpc ndr.Invoker, desiredAccess uint32) (mslsad.LSAPR_HANDLE, error) {
	req := &lsarOpenPolicy2Request{DesiredAccess: ndr.DWORD(desiredAccess)}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mslsad.LSAPR_HANDLE{}, fmt.Errorf("LsarOpenPolicy2: %w", err)
	}
	if nt_status.NT_STATUS(resp.Status) != nt_status.NT_STATUS_SUCCESS {
		return resp.Handle, fmt.Errorf("LsarOpenPolicy2 failed: %s", nt_status.NT_STATUS(resp.Status).String())
	}
	return resp.Handle, nil
}
