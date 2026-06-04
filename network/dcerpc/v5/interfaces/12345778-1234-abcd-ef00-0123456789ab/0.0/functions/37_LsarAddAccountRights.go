package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarAddAccountRightsRequest is the [in] parameter set of LsarAddAccountRights: an open
// policy handle, the SID of the target account, and the inline set of user-right names to
// add (a top-level [ref] struct).
type lsarAddAccountRightsRequest struct {
	PolicyHandle structures.LSAPR_HANDLE
	AccountSid   *dtyp.RPC_SID `ndr:"unique"`
	UserRights   structures.LSAPR_USER_RIGHT_SET
}

func (*lsarAddAccountRightsRequest) Opnum() uint16 { return lsarpc.OpnumLsarAddAccountRights }

// LsarAddAccountRights calls LsarAddAccountRights (opnum 37), granting the named user rights
// to the account identified by SID, creating the account if it does not exist
// ([MS-LSAD] 3.1.4.5.12).
func LsarAddAccountRights(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, accountSid *dtyp.RPC_SID, userRights structures.LSAPR_USER_RIGHT_SET) error {
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
