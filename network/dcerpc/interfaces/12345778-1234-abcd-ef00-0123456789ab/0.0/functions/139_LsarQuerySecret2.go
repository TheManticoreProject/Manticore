package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarQuerySecret2Request carries the [in] parameters of LsarQuerySecret2.
type lsarQuerySecret2Request struct {
	SecretHandle          mslsad.LSAPR_HANDLE
	EncryptedCurrentValue *mslsad.LSAPR_AES_CIPHER_VALUE `ndr:"unique"`
	CurrentValueSetTime   *msdtyp.LARGE_INTEGER          `ndr:"unique"`
	EncryptedOldValue     *mslsad.LSAPR_AES_CIPHER_VALUE `ndr:"unique"`
	OldValueSetTime       *msdtyp.LARGE_INTEGER          `ndr:"unique"`
}

func (*lsarQuerySecret2Request) Opnum() uint16 { return lsarpc.OpnumLsarQuerySecret2 }

// lsarQuerySecret2Response carries the [out] parameters and return value of LsarQuerySecret2.
type lsarQuerySecret2Response struct {
	EncryptedCurrentValue *mslsad.LSAPR_AES_CIPHER_VALUE `ndr:"unique"`
	CurrentValueSetTime   *msdtyp.LARGE_INTEGER          `ndr:"unique"`
	EncryptedOldValue     *mslsad.LSAPR_AES_CIPHER_VALUE `ndr:"unique"`
	OldValueSetTime       *msdtyp.LARGE_INTEGER          `ndr:"unique"`
	Status                ndr.DWORD                      `ndr:"retval"`
}

// LsarQuerySecret2 calls LsarQuerySecret2 (opnum 139) ([MS-LSAD]). Wire
// modeling is not yet live-validated against Windows.
func LsarQuerySecret2(rpc ndr.Invoker, secretHandle mslsad.LSAPR_HANDLE, encryptedCurrentValue *mslsad.LSAPR_AES_CIPHER_VALUE, currentValueSetTime *msdtyp.LARGE_INTEGER, encryptedOldValue *mslsad.LSAPR_AES_CIPHER_VALUE, oldValueSetTime *msdtyp.LARGE_INTEGER) (EncryptedCurrentValue *mslsad.LSAPR_AES_CIPHER_VALUE, CurrentValueSetTime *msdtyp.LARGE_INTEGER, EncryptedOldValue *mslsad.LSAPR_AES_CIPHER_VALUE, OldValueSetTime *msdtyp.LARGE_INTEGER, err error) {
	req := &lsarQuerySecret2Request{
		SecretHandle:          secretHandle,
		EncryptedCurrentValue: encryptedCurrentValue,
		CurrentValueSetTime:   currentValueSetTime,
		EncryptedOldValue:     encryptedOldValue,
		OldValueSetTime:       oldValueSetTime,
	}
	var resp lsarQuerySecret2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LsarQuerySecret2: %w", err)
		return
	}
	EncryptedCurrentValue = resp.EncryptedCurrentValue
	CurrentValueSetTime = resp.CurrentValueSetTime
	EncryptedOldValue = resp.EncryptedOldValue
	OldValueSetTime = resp.OldValueSetTime
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		err = fmt.Errorf("LsarQuerySecret2 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return
}
