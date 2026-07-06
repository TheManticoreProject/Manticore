package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrSetDSRMPasswordRequest carries the [in] parameters. The handle_t binding
// handle is implicit (the RPC client) and is not marshalled; Unused is an ignored
// [unique] string, UserId selects the account, and EncryptedNtOwfPassword is the
// [unique] encrypted NT OWF of the new password.
type samrSetDSRMPasswordRequest struct {
	Unused                 *msdtyp.RPC_UNICODE_STRING `ndr:"unique"`
	UserId                 ndr.DWORD
	EncryptedNtOwfPassword *mssamr.ENCRYPTED_NT_OWF_PASSWORD `ndr:"unique"`
}

func (*samrSetDSRMPasswordRequest) Opnum() uint16 { return samr.OpnumSamrSetDSRMPassword }

// SamrSetDSRMPassword calls SamrSetDSRMPassword (opnum 66), setting the local
// Directory Services Restore Mode administrator password ([MS-SAMR] 3.1.5.13.7).
func SamrSetDSRMPassword(rpc ndr.Invoker, userId uint32, encryptedNtOwfPassword *mssamr.ENCRYPTED_NT_OWF_PASSWORD) error {
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
