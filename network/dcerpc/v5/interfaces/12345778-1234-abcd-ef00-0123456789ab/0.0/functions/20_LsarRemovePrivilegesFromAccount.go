package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarRemovePrivilegesFromAccountRequest is the [in] parameter set of
// LsarRemovePrivilegesFromAccount: an open account handle, the AllPrivileges flag (nonzero
// removes every privilege and ignores Privileges), and a [unique] pointer to the set of
// privileges to remove.
type lsarRemovePrivilegesFromAccountRequest struct {
	AccountHandle structures.LSAPR_HANDLE
	AllPrivileges uint8
	Privileges    *structures.LSAPR_PRIVILEGE_SET `ndr:"unique"`
}

func (*lsarRemovePrivilegesFromAccountRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarRemovePrivilegesFromAccount
}

// LsarRemovePrivilegesFromAccount calls LsarRemovePrivilegesFromAccount (opnum 20), removing
// the given privileges from the account; if allPrivileges is nonzero, every privilege is
// removed and privileges is ignored ([MS-LSAD] 3.1.4.5.6).
func LsarRemovePrivilegesFromAccount(rpc *client.Client, accountHandle structures.LSAPR_HANDLE, allPrivileges uint8, privileges *structures.LSAPR_PRIVILEGE_SET) error {
	req := &lsarRemovePrivilegesFromAccountRequest{
		AccountHandle: accountHandle,
		AllPrivileges: allPrivileges,
		Privileges:    privileges,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarRemovePrivilegesFromAccount: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarRemovePrivilegesFromAccount failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
