package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarOpenAccountRequest is the [in] parameter set of LsarOpenAccount: an open policy
// handle, the SID of the account to open, and the desired access mask.
type lsarOpenAccountRequest struct {
	PolicyHandle  structures.LSAPR_HANDLE
	AccountSid    *dtyp.RPC_SID `ndr:"unique"`
	DesiredAccess ndr.DWORD
}

func (*lsarOpenAccountRequest) Opnum() uint16 { return lsarpc.OpnumLsarOpenAccount }

// LsarOpenAccount calls LsarOpenAccount (opnum 17), opening the account object identified
// by the given SID and returning a handle to it ([MS-LSAD] 3.1.4.5.3).
func LsarOpenAccount(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, accountSid *dtyp.RPC_SID, desiredAccess uint32) (structures.LSAPR_HANDLE, error) {
	req := &lsarOpenAccountRequest{
		PolicyHandle:  policyHandle,
		AccountSid:    accountSid,
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.LSAPR_HANDLE{}, fmt.Errorf("LsarOpenAccount: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarOpenAccount failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
