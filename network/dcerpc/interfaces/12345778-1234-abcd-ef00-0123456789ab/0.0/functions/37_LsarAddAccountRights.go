package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarAddAccountRightsRequest is the [in] parameter set of LsarAddAccountRights: an open
// policy handle, the SID of the target account, and the inline set of user-right names to
// add (a top-level [ref] struct).
type lsarAddAccountRightsRequest struct {
	PolicyHandle mslsad.LSAPR_HANDLE
	AccountSid   *msdtyp.RPC_SID `ndr:"unique"`
	UserRights   mslsad.LSAPR_USER_RIGHT_SET
}

func (*lsarAddAccountRightsRequest) Opnum() uint16 { return lsarpc.OpnumLsarAddAccountRights }

// LsarAddAccountRights calls LsarAddAccountRights (opnum 37), granting the named user rights
// to the account identified by SID, creating the account if it does not exist
// ([MS-LSAD] 3.1.4.5.12).
func LsarAddAccountRights(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, accountSid *msdtyp.RPC_SID, userRights mslsad.LSAPR_USER_RIGHT_SET) error {
	req := &lsarAddAccountRightsRequest{
		PolicyHandle: policyHandle,
		AccountSid:   accountSid,
		UserRights:   userRights,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarAddAccountRights: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarAddAccountRights failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
