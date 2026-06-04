package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarEnumerateAccountRightsRequest is the [in] parameter set of LsarEnumerateAccountRights:
// an open policy handle and the SID of the account whose rights are queried.
type lsarEnumerateAccountRightsRequest struct {
	PolicyHandle structures.LSAPR_HANDLE
	AccountSid   *dtyp.RPC_SID `ndr:"unique"`
}

func (*lsarEnumerateAccountRightsRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarEnumerateAccountRights
}

// lsarEnumerateAccountRightsResponse is the reply: the [out] set of user-right names held by
// the account (an inline [ref] struct) followed by the NTSTATUS return value.
type lsarEnumerateAccountRightsResponse struct {
	UserRights structures.LSAPR_USER_RIGHT_SET
	Status     ndr.DWORD `ndr:"retval"`
}

// LsarEnumerateAccountRights calls LsarEnumerateAccountRights (opnum 36), returning the set
// of user-right names held by the account identified by SID ([MS-LSAD] 3.1.4.5.11).
func LsarEnumerateAccountRights(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, accountSid *dtyp.RPC_SID) (structures.LSAPR_USER_RIGHT_SET, error) {
	req := &lsarEnumerateAccountRightsRequest{PolicyHandle: policyHandle, AccountSid: accountSid}
	var resp lsarEnumerateAccountRightsResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.LSAPR_USER_RIGHT_SET{}, fmt.Errorf("LsarEnumerateAccountRights: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.UserRights, fmt.Errorf("LsarEnumerateAccountRights failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.UserRights, nil
}
