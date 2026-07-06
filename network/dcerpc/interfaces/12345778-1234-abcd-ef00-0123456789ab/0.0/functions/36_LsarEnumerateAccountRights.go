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

// lsarEnumerateAccountRightsRequest is the [in] parameter set of LsarEnumerateAccountRights:
// an open policy handle and the SID of the account whose rights are queried.
type lsarEnumerateAccountRightsRequest struct {
	PolicyHandle mslsad.LSAPR_HANDLE
	AccountSid   *msdtyp.RPC_SID `ndr:"unique"`
}

func (*lsarEnumerateAccountRightsRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarEnumerateAccountRights
}

// lsarEnumerateAccountRightsResponse is the reply: the [out] set of user-right names held by
// the account (an inline [ref] struct) followed by the NTSTATUS return value.
type lsarEnumerateAccountRightsResponse struct {
	UserRights mslsad.LSAPR_USER_RIGHT_SET
	Status     ndr.DWORD `ndr:"retval"`
}

// LsarEnumerateAccountRights calls LsarEnumerateAccountRights (opnum 36), returning the set
// of user-right names held by the account identified by SID ([MS-LSAD] 3.1.4.5.11).
func LsarEnumerateAccountRights(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, accountSid *msdtyp.RPC_SID) (mslsad.LSAPR_USER_RIGHT_SET, error) {
	req := &lsarEnumerateAccountRightsRequest{PolicyHandle: policyHandle, AccountSid: accountSid}
	var resp lsarEnumerateAccountRightsResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mslsad.LSAPR_USER_RIGHT_SET{}, fmt.Errorf("LsarEnumerateAccountRights: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.UserRights, fmt.Errorf("LsarEnumerateAccountRights failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.UserRights, nil
}
