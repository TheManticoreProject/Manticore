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

// lsarRemoveAccountRightsRequest is the [in] parameter set of LsarRemoveAccountRights: an
// open policy handle, the SID of the target account, the AllRights flag (nonzero removes
// every right and ignores UserRights), and the inline set of user-right names to remove
// (a top-level [ref] struct).
type lsarRemoveAccountRightsRequest struct {
	PolicyHandle mslsad.LSAPR_HANDLE
	AccountSid   *msdtyp.RPC_SID `ndr:"unique"`
	AllRights    uint8
	UserRights   mslsad.LSAPR_USER_RIGHT_SET
}

func (*lsarRemoveAccountRightsRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarRemoveAccountRights
}

// LsarRemoveAccountRights calls LsarRemoveAccountRights (opnum 38), revoking the named user
// rights from the account identified by SID; if allRights is nonzero, every right is removed
// and userRights is ignored ([MS-LSAD] 3.1.4.5.13).
func LsarRemoveAccountRights(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, accountSid *msdtyp.RPC_SID, allRights uint8, userRights mslsad.LSAPR_USER_RIGHT_SET) error {
	req := &lsarRemoveAccountRightsRequest{
		PolicyHandle: policyHandle,
		AccountSid:   accountSid,
		AllRights:    allRights,
		UserRights:   userRights,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarRemoveAccountRights: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarRemoveAccountRights failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
