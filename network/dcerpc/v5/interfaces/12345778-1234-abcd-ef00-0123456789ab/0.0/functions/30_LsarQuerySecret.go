package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarQuerySecretRequest is the [in,out] parameter set of LsarQuerySecret: an open secret
// handle plus four [in,out,unique] slots the client passes (typically as NULL referents)
// to select which values the server should return. The double-pointer cipher-value
// parameters become a [unique] pointer to the inner [unique] LSAPR_CR_CIPHER_VALUE, and
// the set-time parameters become [unique] pointers to a LARGE_INTEGER.
type lsarQuerySecretRequest struct {
	SecretHandle          structures.LSAPR_HANDLE
	EncryptedCurrentValue *structures.LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
	CurrentValueSetTime   *dtyp.LARGE_INTEGER               `ndr:"unique"`
	EncryptedOldValue     *structures.LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
	OldValueSetTime       *dtyp.LARGE_INTEGER               `ndr:"unique"`
}

func (*lsarQuerySecretRequest) Opnum() uint16 { return lsarpc.OpnumLsarQuerySecret }

// lsarQuerySecretResponse is the reply: the four [in,out,unique] values populated by the
// server followed by the NTSTATUS return value.
type lsarQuerySecretResponse struct {
	EncryptedCurrentValue *structures.LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
	CurrentValueSetTime   *dtyp.LARGE_INTEGER               `ndr:"unique"`
	EncryptedOldValue     *structures.LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
	OldValueSetTime       *dtyp.LARGE_INTEGER               `ndr:"unique"`
	Status                ndr.DWORD                         `ndr:"retval"`
}

// LsarQuerySecret calls LsarQuerySecret (opnum 30) and returns the current and old
// encrypted values of the secret together with their set times ([MS-LSAD] 3.1.4.6.5).
// Any output may be nil if the server did not return that value.
func LsarQuerySecret(rpc *client.Client, secretHandle structures.LSAPR_HANDLE) (encryptedCurrentValue *structures.LSAPR_CR_CIPHER_VALUE, currentValueSetTime *dtyp.LARGE_INTEGER, encryptedOldValue *structures.LSAPR_CR_CIPHER_VALUE, oldValueSetTime *dtyp.LARGE_INTEGER, err error) {
	req := &lsarQuerySecretRequest{SecretHandle: secretHandle}
	var resp lsarQuerySecretResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("LsarQuerySecret: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.EncryptedCurrentValue, resp.CurrentValueSetTime, resp.EncryptedOldValue, resp.OldValueSetTime, fmt.Errorf("LsarQuerySecret failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.EncryptedCurrentValue, resp.CurrentValueSetTime, resp.EncryptedOldValue, resp.OldValueSetTime, nil
}
