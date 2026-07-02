package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarStorePrivateData2Request carries the [in] parameters of LsarStorePrivateData2.
type lsarStorePrivateData2Request struct {
	PolicyHandle     mslsad.LSAPR_HANDLE
	EncryptedKeyName mslsad.LSAPR_AES_CIPHER_VALUE
	EncryptedData    *mslsad.LSAPR_AES_CIPHER_VALUE `ndr:"unique"`
}

func (*lsarStorePrivateData2Request) Opnum() uint16 { return lsarpc.OpnumLsarStorePrivateData2 }

// lsarStorePrivateData2Response carries the [out] parameters and return value of LsarStorePrivateData2.
type lsarStorePrivateData2Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// LsarStorePrivateData2 calls LsarStorePrivateData2 (opnum 140) ([MS-LSAD]). Wire
// modeling is not yet live-validated against Windows.
func LsarStorePrivateData2(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, encryptedKeyName mslsad.LSAPR_AES_CIPHER_VALUE, encryptedData *mslsad.LSAPR_AES_CIPHER_VALUE) (err error) {
	req := &lsarStorePrivateData2Request{
		PolicyHandle:     policyHandle,
		EncryptedKeyName: encryptedKeyName,
		EncryptedData:    encryptedData,
	}
	var resp lsarStorePrivateData2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LsarStorePrivateData2: %w", err)
		return
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		err = fmt.Errorf("LsarStorePrivateData2 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return
}
