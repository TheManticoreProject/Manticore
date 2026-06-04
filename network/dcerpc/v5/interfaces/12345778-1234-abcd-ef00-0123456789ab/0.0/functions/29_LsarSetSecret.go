package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarSetSecretRequest is the [in] parameter set of LsarSetSecret: an open secret handle
// and two [unique] encrypted values (current and old). A nil pointer marshals as a NULL
// referent, which the server interprets per [MS-LSAD] 3.1.4.6.4 (e.g. clearing a value).
type lsarSetSecretRequest struct {
	SecretHandle          structures.LSAPR_HANDLE
	EncryptedCurrentValue *structures.LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
	EncryptedOldValue     *structures.LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
}

func (*lsarSetSecretRequest) Opnum() uint16 { return lsarpc.OpnumLsarSetSecret }

// LsarSetSecret calls LsarSetSecret (opnum 29), setting the current and old encrypted
// values of the secret referenced by the handle ([MS-LSAD] 3.1.4.6.4).
func LsarSetSecret(rpc *client.Client, secretHandle structures.LSAPR_HANDLE, encryptedCurrentValue, encryptedOldValue *structures.LSAPR_CR_CIPHER_VALUE) error {
	req := &lsarSetSecretRequest{
		SecretHandle:          secretHandle,
		EncryptedCurrentValue: encryptedCurrentValue,
		EncryptedOldValue:     encryptedOldValue,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarSetSecret: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarSetSecret failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
