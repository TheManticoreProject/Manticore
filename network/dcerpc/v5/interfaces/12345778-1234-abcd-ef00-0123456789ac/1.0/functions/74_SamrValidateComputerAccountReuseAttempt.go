package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrValidateComputerAccountReuseAttemptRequest carries the [in] server handle and the
// [ref] SID (inline, single pointer) of the computer account whose reuse is being validated.
type samrValidateComputerAccountReuseAttemptRequest struct {
	ServerHandle structures.SAMPR_HANDLE
	ComputerSid  dtyp.RPC_SID
}

func (*samrValidateComputerAccountReuseAttemptRequest) Opnum() uint16 {
	return samr.OpnumSamrValidateComputerAccountReuseAttempt
}

// samrValidateComputerAccountReuseAttemptResponse is the reply: the [out] BOOL result (a
// 32-bit value) and the NTSTATUS.
type samrValidateComputerAccountReuseAttemptResponse struct {
	Result int32
	Status ndr.DWORD `ndr:"retval"`
}

// SamrValidateComputerAccountReuseAttempt calls SamrValidateComputerAccountReuseAttempt
// (opnum 74), determining whether the caller is allowed to reuse the given computer account
// SID ([MS-SAMR] 3.1.5.13.5).
func SamrValidateComputerAccountReuseAttempt(rpc *client.Client, serverHandle structures.SAMPR_HANDLE, computerSid dtyp.RPC_SID) (int32, error) {
	req := &samrValidateComputerAccountReuseAttemptRequest{
		ServerHandle: serverHandle,
		ComputerSid:  computerSid,
	}
	var resp samrValidateComputerAccountReuseAttemptResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, fmt.Errorf("SamrValidateComputerAccountReuseAttempt: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Result, fmt.Errorf("SamrValidateComputerAccountReuseAttempt failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Result, nil
}
