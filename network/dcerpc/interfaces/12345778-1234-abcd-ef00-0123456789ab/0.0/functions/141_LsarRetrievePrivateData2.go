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

// lsarRetrievePrivateData2Request carries the [in] parameters of LsarRetrievePrivateData2.
type lsarRetrievePrivateData2Request struct {
	PolicyHandle     mslsad.LSAPR_HANDLE
	EncryptedKeyName mslsad.LSAPR_AES_CIPHER_VALUE
	EncryptedData    *mslsad.LSAPR_AES_CIPHER_VALUE `ndr:"unique"`
}

func (*lsarRetrievePrivateData2Request) Opnum() uint16 { return lsarpc.OpnumLsarRetrievePrivateData2 }

// lsarRetrievePrivateData2Response carries the [out] parameters and return value of LsarRetrievePrivateData2.
type lsarRetrievePrivateData2Response struct {
	EncryptedData *mslsad.LSAPR_AES_CIPHER_VALUE `ndr:"unique"`
	Status        ndr.DWORD                      `ndr:"retval"`
}

// LsarRetrievePrivateData2 calls LsarRetrievePrivateData2 (opnum 141) ([MS-LSAD]). Wire
// modeling is not yet live-validated against Windows.
func LsarRetrievePrivateData2(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, encryptedKeyName mslsad.LSAPR_AES_CIPHER_VALUE, encryptedData *mslsad.LSAPR_AES_CIPHER_VALUE) (EncryptedData *mslsad.LSAPR_AES_CIPHER_VALUE, err error) {
	req := &lsarRetrievePrivateData2Request{
		PolicyHandle:     policyHandle,
		EncryptedKeyName: encryptedKeyName,
		EncryptedData:    encryptedData,
	}
	var resp lsarRetrievePrivateData2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LsarRetrievePrivateData2: %w", err)
		return
	}
	EncryptedData = resp.EncryptedData
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		err = fmt.Errorf("LsarRetrievePrivateData2 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return
}
