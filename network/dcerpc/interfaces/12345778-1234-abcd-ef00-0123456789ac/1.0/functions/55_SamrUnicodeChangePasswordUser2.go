package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrUnicodeChangePasswordUser2Request carries the [in] parameters of
// SamrUnicodeChangePasswordUser2 (the handle_t binding handle is implicit and omitted): the
// [unique] Unicode server name, the [ref] Unicode user name, the NT-encrypted new password
// and old NT OWF, an LM-present flag, and the optional LM-encrypted new password and old LM
// OWF. Field order matches the IDL exactly.
type samrUnicodeChangePasswordUser2Request struct {
	ServerName                         *msdtyp.RPC_UNICODE_STRING `ndr:"unique"`
	UserName                           msdtyp.RPC_UNICODE_STRING
	NewPasswordEncryptedWithOldNt      *mssamr.SAMPR_ENCRYPTED_USER_PASSWORD `ndr:"unique"`
	OldNtOwfPasswordEncryptedWithNewNt *mssamr.ENCRYPTED_NT_OWF_PASSWORD     `ndr:"unique"`
	LmPresent                          uint8
	NewPasswordEncryptedWithOldLm      *mssamr.SAMPR_ENCRYPTED_USER_PASSWORD `ndr:"unique"`
	OldLmOwfPasswordEncryptedWithNewNt *mssamr.ENCRYPTED_LM_OWF_PASSWORD     `ndr:"unique"`
}

func (*samrUnicodeChangePasswordUser2Request) Opnum() uint16 {
	return samr.OpnumSamrUnicodeChangePasswordUser2
}

// SamrUnicodeChangePasswordUser2 calls SamrUnicodeChangePasswordUser2 (opnum 55), changing a
// user's password using Unicode names and NT-based encryption (with optional LM material)
// ([MS-SAMR] 3.1.5.10.4).
func SamrUnicodeChangePasswordUser2(rpc ndr.Invoker, serverName *msdtyp.RPC_UNICODE_STRING, userName msdtyp.RPC_UNICODE_STRING, newPasswordEncryptedWithOldNt *mssamr.SAMPR_ENCRYPTED_USER_PASSWORD, oldNtOwfPasswordEncryptedWithNewNt *mssamr.ENCRYPTED_NT_OWF_PASSWORD, lmPresent uint8, newPasswordEncryptedWithOldLm *mssamr.SAMPR_ENCRYPTED_USER_PASSWORD, oldLmOwfPasswordEncryptedWithNewNt *mssamr.ENCRYPTED_LM_OWF_PASSWORD) error {
	req := &samrUnicodeChangePasswordUser2Request{
		ServerName:                         serverName,
		UserName:                           userName,
		NewPasswordEncryptedWithOldNt:      newPasswordEncryptedWithOldNt,
		OldNtOwfPasswordEncryptedWithNewNt: oldNtOwfPasswordEncryptedWithNewNt,
		LmPresent:                          lmPresent,
		NewPasswordEncryptedWithOldLm:      newPasswordEncryptedWithOldLm,
		OldLmOwfPasswordEncryptedWithNewNt: oldLmOwfPasswordEncryptedWithNewNt,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrUnicodeChangePasswordUser2: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrUnicodeChangePasswordUser2 failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
