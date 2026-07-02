package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarSetSecret2Request carries the [in] parameters of LsarSetSecret2.
type lsarSetSecret2Request struct {
	SecretHandle          mslsad.LSAPR_HANDLE
	EncryptedCurrentValue *mslsad.LSAPR_AES_CIPHER_VALUE `ndr:"unique"`
	EncryptedOldValue     *mslsad.LSAPR_AES_CIPHER_VALUE `ndr:"unique"`
}

func (*lsarSetSecret2Request) Opnum() uint16 { return lsarpc.OpnumLsarSetSecret2 }

// lsarSetSecret2Response carries the [out] parameters and return value of LsarSetSecret2.
type lsarSetSecret2Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// LsarSetSecret2 calls LsarSetSecret2 (opnum 138) ([MS-LSAD]). Wire
// modeling is not yet live-validated against Windows.
func LsarSetSecret2(rpc ndr.Invoker, secretHandle mslsad.LSAPR_HANDLE, encryptedCurrentValue *mslsad.LSAPR_AES_CIPHER_VALUE, encryptedOldValue *mslsad.LSAPR_AES_CIPHER_VALUE) (err error) {
	req := &lsarSetSecret2Request{
		SecretHandle:          secretHandle,
		EncryptedCurrentValue: encryptedCurrentValue,
		EncryptedOldValue:     encryptedOldValue,
	}
	var resp lsarSetSecret2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LsarSetSecret2: %w", err)
		return
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		err = fmt.Errorf("LsarSetSecret2 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return
}
