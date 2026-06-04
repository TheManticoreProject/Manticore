package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarRemoveAccountRightsRequest is the [in] parameter set of LsarRemoveAccountRights: an
// open policy handle, the SID of the target account, the AllRights flag (nonzero removes
// every right and ignores UserRights), and the inline set of user-right names to remove
// (a top-level [ref] struct).
type lsarRemoveAccountRightsRequest struct {
	PolicyHandle structures.LSAPR_HANDLE
	AccountSid   *dtyp.RPC_SID `ndr:"unique"`
	AllRights    uint8
	UserRights   structures.LSAPR_USER_RIGHT_SET
}

func (*lsarRemoveAccountRightsRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarRemoveAccountRights
}

// LsarRemoveAccountRights calls LsarRemoveAccountRights (opnum 38), revoking the named user
// rights from the account identified by SID; if allRights is nonzero, every right is removed
// and userRights is ignored ([MS-LSAD] 3.1.4.5.13).
func LsarRemoveAccountRights(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, accountSid *dtyp.RPC_SID, allRights uint8, userRights structures.LSAPR_USER_RIGHT_SET) error {
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
