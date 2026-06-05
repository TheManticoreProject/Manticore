package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrOemChangePasswordUser2Request carries the [in] parameters of SamrOemChangePasswordUser2
// (the handle_t binding handle is implicit and omitted): the [unique] ANSI server name, the
// [ref] ANSI user name, and the new password / old LM OWF blobs cross-encrypted with one
// another. Field order matches the IDL.
type samrOemChangePasswordUser2Request struct {
	ServerName                         *structures.RPC_STRING `ndr:"unique"`
	UserName                           structures.RPC_STRING
	NewPasswordEncryptedWithOldLm      *structures.SAMPR_ENCRYPTED_USER_PASSWORD `ndr:"unique"`
	OldLmOwfPasswordEncryptedWithNewLm *structures.ENCRYPTED_LM_OWF_PASSWORD     `ndr:"unique"`
}

func (*samrOemChangePasswordUser2Request) Opnum() uint16 {
	return samr.OpnumSamrOemChangePasswordUser2
}

// SamrOemChangePasswordUser2 calls SamrOemChangePasswordUser2 (opnum 54), changing a user's
// password using OEM (ANSI) encoded names and LM-based encryption ([MS-SAMR] 3.1.5.10.3).
func SamrOemChangePasswordUser2(rpc ndr.Invoker, serverName *structures.RPC_STRING, userName structures.RPC_STRING, newPasswordEncryptedWithOldLm *structures.SAMPR_ENCRYPTED_USER_PASSWORD, oldLmOwfPasswordEncryptedWithNewLm *structures.ENCRYPTED_LM_OWF_PASSWORD) error {
	req := &samrOemChangePasswordUser2Request{
		ServerName:                         serverName,
		UserName:                           userName,
		NewPasswordEncryptedWithOldLm:      newPasswordEncryptedWithOldLm,
		OldLmOwfPasswordEncryptedWithNewLm: oldLmOwfPasswordEncryptedWithNewLm,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrOemChangePasswordUser2: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrOemChangePasswordUser2 failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
