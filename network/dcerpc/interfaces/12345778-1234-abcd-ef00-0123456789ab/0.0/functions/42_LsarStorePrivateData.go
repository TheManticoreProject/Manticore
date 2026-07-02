package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarStorePrivateDataRequest is the [in] parameter set of LsarStorePrivateData: an open
// policy handle, the key name (a [ref] PRPC_UNICODE_STRING, modeled inline), and the
// [unique] encrypted data. A nil EncryptedData marshals as a NULL referent, which the
// server interprets as a request to delete the value ([MS-LSAD] 3.1.4.7.1).
type lsarStorePrivateDataRequest struct {
	PolicyHandle  mslsad.LSAPR_HANDLE
	KeyName       dtyp.RPC_UNICODE_STRING
	EncryptedData *mslsad.LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
}

func (*lsarStorePrivateDataRequest) Opnum() uint16 { return lsarpc.OpnumLsarStorePrivateData }

// LsarStorePrivateData calls LsarStorePrivateData (opnum 42), storing the encrypted data
// under the given key name ([MS-LSAD] 3.1.4.7.1).
func LsarStorePrivateData(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, keyName string, encryptedData *mslsad.LSAPR_CR_CIPHER_VALUE) error {
	req := &lsarStorePrivateDataRequest{
		PolicyHandle:  policyHandle,
		KeyName:       dtyp.NewUnicodeString(keyName),
		EncryptedData: encryptedData,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarStorePrivateData: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarStorePrivateData failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
