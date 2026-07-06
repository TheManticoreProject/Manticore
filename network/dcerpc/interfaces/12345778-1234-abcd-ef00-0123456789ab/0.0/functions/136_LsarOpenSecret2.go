package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarOpenSecret2Request carries the [in] parameters of LsarOpenSecret2.
type lsarOpenSecret2Request struct {
	PolicyHandle        mslsad.LSAPR_HANDLE
	EncryptedSecretName mslsad.LSAPR_AES_CIPHER_VALUE
	DesiredAccess       ndr.DWORD
}

func (*lsarOpenSecret2Request) Opnum() uint16 { return lsarpc.OpnumLsarOpenSecret2 }

// lsarOpenSecret2Response carries the [out] parameters and return value of LsarOpenSecret2.
type lsarOpenSecret2Response struct {
	SecretHandle mslsad.LSAPR_HANDLE
	Status       ndr.DWORD `ndr:"retval"`
}

// LsarOpenSecret2 calls LsarOpenSecret2 (opnum 136) ([MS-LSAD]). Wire
// modeling is not yet live-validated against Windows.
func LsarOpenSecret2(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, encryptedSecretName mslsad.LSAPR_AES_CIPHER_VALUE, desiredAccess ndr.DWORD) (SecretHandle mslsad.LSAPR_HANDLE, err error) {
	req := &lsarOpenSecret2Request{
		PolicyHandle:        policyHandle,
		EncryptedSecretName: encryptedSecretName,
		DesiredAccess:       desiredAccess,
	}
	var resp lsarOpenSecret2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LsarOpenSecret2: %w", err)
		return
	}
	SecretHandle = resp.SecretHandle
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		err = fmt.Errorf("LsarOpenSecret2 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return
}
