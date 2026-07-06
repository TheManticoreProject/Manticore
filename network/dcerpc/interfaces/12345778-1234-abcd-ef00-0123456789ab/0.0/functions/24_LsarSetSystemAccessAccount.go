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

// lsarSetSystemAccessAccountRequest is the [in] parameter set of
// LsarSetSystemAccessAccount: an open account handle and the new system-access flags.
type lsarSetSystemAccessAccountRequest struct {
	AccountHandle mslsad.LSAPR_HANDLE
	SystemAccess  ndr.DWORD
}

func (*lsarSetSystemAccessAccountRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarSetSystemAccessAccount
}

// LsarSetSystemAccessAccount calls LsarSetSystemAccessAccount (opnum 24), setting the
// system-access flags of the account.
func LsarSetSystemAccessAccount(rpc ndr.Invoker, accountHandle mslsad.LSAPR_HANDLE, systemAccess uint32) error {
	req := &lsarSetSystemAccessAccountRequest{AccountHandle: accountHandle, SystemAccess: ndr.DWORD(systemAccess)}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarSetSystemAccessAccount: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarSetSystemAccessAccount failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
