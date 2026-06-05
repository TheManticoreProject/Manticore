package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// lsarEnumeratePrivilegesAccountRequest is the [in] parameter of
// LsarEnumeratePrivilegesAccount: an open account handle.
type lsarEnumeratePrivilegesAccountRequest struct {
	AccountHandle structures.LSAPR_HANDLE
}

func (*lsarEnumeratePrivilegesAccountRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarEnumeratePrivilegesAccount
}

// lsarEnumeratePrivilegesAccountResponse is the reply: the [out] set of privileges held by
// the account (a [unique] double pointer) followed by the NTSTATUS return value.
type lsarEnumeratePrivilegesAccountResponse struct {
	Privileges *structures.LSAPR_PRIVILEGE_SET `ndr:"unique"`
	Status     ndr.DWORD                       `ndr:"retval"`
}

// LsarEnumeratePrivilegesAccount calls LsarEnumeratePrivilegesAccount (opnum 18), returning
// the set of privileges held by the account ([MS-LSAD] 3.1.4.5.4).
func LsarEnumeratePrivilegesAccount(rpc ndr.Invoker, accountHandle structures.LSAPR_HANDLE) (*structures.LSAPR_PRIVILEGE_SET, error) {
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
