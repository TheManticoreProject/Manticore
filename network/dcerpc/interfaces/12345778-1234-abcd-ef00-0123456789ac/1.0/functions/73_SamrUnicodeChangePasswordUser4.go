package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrUnicodeChangePasswordUser4Request carries the [in] parameters of
// SamrUnicodeChangePasswordUser4 (the handle_t binding handle is implicit and omitted): the
// [unique] Unicode server name, the [ref] Unicode user name, and the AES-encrypted password
// blob (inline, single [ref] pointer).
type samrUnicodeChangePasswordUser4Request struct {
	ServerName        *msdtyp.RPC_UNICODE_STRING `ndr:"unique"`
	UserName          msdtyp.RPC_UNICODE_STRING
	EncryptedPassword mssamr.SAMPR_ENCRYPTED_PASSWORD_AES
}

func (*samrUnicodeChangePasswordUser4Request) Opnum() uint16 {
	return samr.OpnumSamrUnicodeChangePasswordUser4
}

// SamrUnicodeChangePasswordUser4 calls SamrUnicodeChangePasswordUser4 (opnum 73), changing a
// user's password using an AES-encrypted password blob ([MS-SAMR] 3.1.5.10.5).
func SamrUnicodeChangePasswordUser4(rpc ndr.Invoker, serverName *msdtyp.RPC_UNICODE_STRING, userName msdtyp.RPC_UNICODE_STRING, encryptedPassword mssamr.SAMPR_ENCRYPTED_PASSWORD_AES) error {
	req := &samrUnicodeChangePasswordUser4Request{
		ServerName:        serverName,
		UserName:          userName,
		EncryptedPassword: encryptedPassword,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrUnicodeChangePasswordUser4: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrUnicodeChangePasswordUser4 failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
