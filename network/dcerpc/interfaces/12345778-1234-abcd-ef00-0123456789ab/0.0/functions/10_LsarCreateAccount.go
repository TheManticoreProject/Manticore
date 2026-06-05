package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// lsarCreateAccountRequest is the [in] parameter set of LsarCreateAccount: an open policy
// handle, the SID of the account to create, and the desired access mask.
type lsarCreateAccountRequest struct {
	PolicyHandle  structures.LSAPR_HANDLE
	AccountSid    *dtyp.RPC_SID `ndr:"unique"`
	DesiredAccess ndr.DWORD
}

func (*lsarCreateAccountRequest) Opnum() uint16 { return lsarpc.OpnumLsarCreateAccount }

// LsarCreateAccount calls LsarCreateAccount (opnum 10), creating an account object for the
// given SID and returning a handle to it ([MS-LSAD] 3.1.4.5.1).
func LsarCreateAccount(rpc ndr.Invoker, policyHandle structures.LSAPR_HANDLE, accountSid *dtyp.RPC_SID, desiredAccess uint32) (structures.LSAPR_HANDLE, error) {
	req := &lsarCreateAccountRequest{
		PolicyHandle:  policyHandle,
		AccountSid:    accountSid,
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.LSAPR_HANDLE{}, fmt.Errorf("LsarCreateAccount: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarCreateAccount failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
