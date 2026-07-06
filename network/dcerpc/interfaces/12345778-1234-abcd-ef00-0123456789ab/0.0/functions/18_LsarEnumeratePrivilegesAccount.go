package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarEnumeratePrivilegesAccountRequest is the [in] parameter of
// LsarEnumeratePrivilegesAccount: an open account handle.
type lsarEnumeratePrivilegesAccountRequest struct {
	AccountHandle mslsad.LSAPR_HANDLE
}

func (*lsarEnumeratePrivilegesAccountRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarEnumeratePrivilegesAccount
}

// lsarEnumeratePrivilegesAccountResponse is the reply: the [out] set of privileges held by
// the account (a [unique] double pointer) followed by the NTSTATUS return value.
type lsarEnumeratePrivilegesAccountResponse struct {
	Privileges *mslsad.LSAPR_PRIVILEGE_SET `ndr:"unique"`
	Status     ndr.DWORD                   `ndr:"retval"`
}

// LsarEnumeratePrivilegesAccount calls LsarEnumeratePrivilegesAccount (opnum 18), returning
// the set of privileges held by the account ([MS-LSAD] 3.1.4.5.4).
func LsarEnumeratePrivilegesAccount(rpc ndr.Invoker, accountHandle mslsad.LSAPR_HANDLE) (*mslsad.LSAPR_PRIVILEGE_SET, error) {
	req := &lsarEnumeratePrivilegesAccountRequest{AccountHandle: accountHandle}
	var resp lsarEnumeratePrivilegesAccountResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("LsarEnumeratePrivilegesAccount: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Privileges, fmt.Errorf("LsarEnumeratePrivilegesAccount failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Privileges, nil
}
