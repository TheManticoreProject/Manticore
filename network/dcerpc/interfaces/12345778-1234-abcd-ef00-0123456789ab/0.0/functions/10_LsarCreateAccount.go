package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarCreateAccountRequest is the [in] parameter set of LsarCreateAccount: an open policy
// handle, the SID of the account to create, and the desired access mask.
type lsarCreateAccountRequest struct {
	PolicyHandle  mslsad.LSAPR_HANDLE
	AccountSid    *msdtyp.RPC_SID `ndr:"unique"`
	DesiredAccess ndr.DWORD
}

func (*lsarCreateAccountRequest) Opnum() uint16 { return lsarpc.OpnumLsarCreateAccount }

// LsarCreateAccount calls LsarCreateAccount (opnum 10), creating an account object for the
// given SID and returning a handle to it ([MS-LSAD] 3.1.4.5.1).
func LsarCreateAccount(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, accountSid *msdtyp.RPC_SID, desiredAccess uint32) (mslsad.LSAPR_HANDLE, error) {
	req := &lsarCreateAccountRequest{
		PolicyHandle:  policyHandle,
		AccountSid:    accountSid,
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mslsad.LSAPR_HANDLE{}, fmt.Errorf("LsarCreateAccount: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarCreateAccount failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
