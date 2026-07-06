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

// lsarGetSystemAccessAccountRequest is the [in] parameter of LsarGetSystemAccessAccount:
// an open account handle.
type lsarGetSystemAccessAccountRequest struct {
	AccountHandle mslsad.LSAPR_HANDLE
}

func (*lsarGetSystemAccessAccountRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarGetSystemAccessAccount
}

// lsarGetSystemAccessAccountResponse is the reply: the [out] system-access flags
// followed by the NTSTATUS return value.
type lsarGetSystemAccessAccountResponse struct {
	SystemAccess ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// LsarGetSystemAccessAccount calls LsarGetSystemAccessAccount (opnum 23) and returns the
// system-access flags of the account ([MS-LSAD] 2.2.1.2 ACCESS_MASK for system access).
func LsarGetSystemAccessAccount(rpc ndr.Invoker, accountHandle mslsad.LSAPR_HANDLE) (uint32, error) {
	req := &lsarGetSystemAccessAccountRequest{AccountHandle: accountHandle}
	var resp lsarGetSystemAccessAccountResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, fmt.Errorf("LsarGetSystemAccessAccount: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return 0, fmt.Errorf("LsarGetSystemAccessAccount failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return uint32(resp.SystemAccess), nil
}
