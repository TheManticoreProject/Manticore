package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarCreateSecret2Request carries the [in] parameters of LsarCreateSecret2.
type lsarCreateSecret2Request struct {
	PolicyHandle        mslsad.LSAPR_HANDLE
	EncryptedSecretName mslsad.LSAPR_AES_CIPHER_VALUE
	DesiredAccess       ndr.DWORD
}

func (*lsarCreateSecret2Request) Opnum() uint16 { return lsarpc.OpnumLsarCreateSecret2 }

// lsarCreateSecret2Response carries the [out] parameters and return value of LsarCreateSecret2.
type lsarCreateSecret2Response struct {
	SecretHandle mslsad.LSAPR_HANDLE
	Status       ndr.DWORD `ndr:"retval"`
}

// LsarCreateSecret2 calls LsarCreateSecret2 (opnum 137) ([MS-LSAD]). Wire
// modeling is not yet live-validated against Windows.
func LsarCreateSecret2(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, encryptedSecretName mslsad.LSAPR_AES_CIPHER_VALUE, desiredAccess ndr.DWORD) (SecretHandle mslsad.LSAPR_HANDLE, err error) {
	req := &lsarCreateSecret2Request{
		PolicyHandle:        policyHandle,
		EncryptedSecretName: encryptedSecretName,
		DesiredAccess:       desiredAccess,
	}
	var resp lsarCreateSecret2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LsarCreateSecret2: %w", err)
		return
	}
	SecretHandle = resp.SecretHandle
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		err = fmt.Errorf("LsarCreateSecret2 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return
}
