package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarSetSecretRequest is the [in] parameter set of LsarSetSecret: an open secret handle
// and two [unique] encrypted values (current and old). A nil pointer marshals as a NULL
// referent, which the server interprets per [MS-LSAD] 3.1.4.6.4 (e.g. clearing a value).
type lsarSetSecretRequest struct {
	SecretHandle          mslsad.LSAPR_HANDLE
	EncryptedCurrentValue *mslsad.LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
	EncryptedOldValue     *mslsad.LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
}

func (*lsarSetSecretRequest) Opnum() uint16 { return lsarpc.OpnumLsarSetSecret }

// LsarSetSecret calls LsarSetSecret (opnum 29), setting the current and old encrypted
// values of the secret referenced by the handle ([MS-LSAD] 3.1.4.6.4).
func LsarSetSecret(rpc ndr.Invoker, secretHandle mslsad.LSAPR_HANDLE, encryptedCurrentValue, encryptedOldValue *mslsad.LSAPR_CR_CIPHER_VALUE) error {
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
