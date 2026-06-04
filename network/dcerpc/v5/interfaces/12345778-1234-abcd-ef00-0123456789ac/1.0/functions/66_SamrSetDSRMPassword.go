package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrSetDSRMPasswordRequest carries the [in] parameters. The handle_t binding
// handle is implicit (the RPC client) and is not marshalled; Unused is an ignored
// [unique] string, UserId selects the account, and EncryptedNtOwfPassword is the
// [unique] encrypted NT OWF of the new password.
type samrSetDSRMPasswordRequest struct {
	Unused                 *dtyp.RPC_UNICODE_STRING `ndr:"unique"`
	UserId                 ndr.DWORD
	EncryptedNtOwfPassword *structures.ENCRYPTED_NT_OWF_PASSWORD `ndr:"unique"`
}

func (*samrSetDSRMPasswordRequest) Opnum() uint16 { return samr.OpnumSamrSetDSRMPassword }

// SamrSetDSRMPassword calls SamrSetDSRMPassword (opnum 66), setting the local
// Directory Services Restore Mode administrator password ([MS-SAMR] 3.1.5.13.7).
func SamrSetDSRMPassword(rpc *client.Client, userId uint32, encryptedNtOwfPassword *structures.ENCRYPTED_NT_OWF_PASSWORD) error {
	req := &samrSetDSRMPasswordRequest{
		UserId:                 ndr.DWORD(userId),
		EncryptedNtOwfPassword: encryptedNtOwfPassword,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrSetDSRMPassword: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrSetDSRMPassword failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
