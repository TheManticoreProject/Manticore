package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrUnicodeChangePasswordUser4Request carries the [in] parameters of
// SamrUnicodeChangePasswordUser4 (the handle_t binding handle is implicit and omitted): the
// [unique] Unicode server name, the [ref] Unicode user name, and the AES-encrypted password
// blob (inline, single [ref] pointer).
type samrUnicodeChangePasswordUser4Request struct {
	ServerName        *dtyp.RPC_UNICODE_STRING `ndr:"unique"`
	UserName          dtyp.RPC_UNICODE_STRING
	EncryptedPassword structures.SAMPR_ENCRYPTED_PASSWORD_AES
}

func (*samrUnicodeChangePasswordUser4Request) Opnum() uint16 {
	return samr.OpnumSamrUnicodeChangePasswordUser4
}

// SamrUnicodeChangePasswordUser4 calls SamrUnicodeChangePasswordUser4 (opnum 73), changing a
// user's password using an AES-encrypted password blob ([MS-SAMR] 3.1.5.10.5).
func SamrUnicodeChangePasswordUser4(rpc *client.Client, serverName *dtyp.RPC_UNICODE_STRING, userName dtyp.RPC_UNICODE_STRING, encryptedPassword structures.SAMPR_ENCRYPTED_PASSWORD_AES) error {
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
