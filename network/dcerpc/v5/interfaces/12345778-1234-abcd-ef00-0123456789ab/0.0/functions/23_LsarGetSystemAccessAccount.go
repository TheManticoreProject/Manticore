package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarGetSystemAccessAccountRequest is the [in] parameter of LsarGetSystemAccessAccount:
// an open account handle.
type lsarGetSystemAccessAccountRequest struct {
	AccountHandle structures.LSAPR_HANDLE
}

func (*lsarGetSystemAccessAccountRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarGetSystemAccessAccount
}

// lsarGetSystemAccessAccountResponse is the reply: the [out] system-access flags
// followed by the NTSTATUS return value.
type lsarGetSystemAccessAccountResponse struct {
	SystemAccess ndr.DWORD
	Status       ndr.DWORD
}

// LsarGetSystemAccessAccount calls LsarGetSystemAccessAccount (opnum 23) and returns the
// system-access flags of the account ([MS-LSAD] 2.2.1.2 ACCESS_MASK for system access).
func LsarGetSystemAccessAccount(rpc *client.Client, accountHandle structures.LSAPR_HANDLE) (uint32, error) {
	req := &lsarGetSystemAccessAccountRequest{AccountHandle: accountHandle}
	var resp lsarGetSystemAccessAccountResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, fmt.Errorf("LsarGetSystemAccessAccount: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return 0, fmt.Errorf("LsarGetSystemAccessAccount failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return uint32(resp.SystemAccess), nil
}
