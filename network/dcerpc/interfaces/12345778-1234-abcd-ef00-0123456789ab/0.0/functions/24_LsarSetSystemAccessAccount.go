package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// lsarSetSystemAccessAccountRequest is the [in] parameter set of
// LsarSetSystemAccessAccount: an open account handle and the new system-access flags.
type lsarSetSystemAccessAccountRequest struct {
	AccountHandle structures.LSAPR_HANDLE
	SystemAccess  ndr.DWORD
}

func (*lsarSetSystemAccessAccountRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarSetSystemAccessAccount
}

// LsarSetSystemAccessAccount calls LsarSetSystemAccessAccount (opnum 24), setting the
// system-access flags of the account.
func LsarSetSystemAccessAccount(rpc ndr.Invoker, accountHandle structures.LSAPR_HANDLE, systemAccess uint32) error {
	req := &lsarSetSystemAccessAccountRequest{AccountHandle: accountHandle, SystemAccess: ndr.DWORD(systemAccess)}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarSetSystemAccessAccount: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarSetSystemAccessAccount failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
