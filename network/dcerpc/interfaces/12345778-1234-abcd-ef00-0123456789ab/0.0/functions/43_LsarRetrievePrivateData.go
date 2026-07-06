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

// lsarRetrievePrivateDataRequest is the [in,out] parameter set of LsarRetrievePrivateData:
// an open policy handle, the key name (a [ref] PRPC_UNICODE_STRING, modeled inline), and
// the [in,out] EncryptedData slot. The double-pointer EncryptedData becomes a [unique]
// pointer to the inner [unique] LSAPR_CR_CIPHER_VALUE; the client passes it as a NULL
// referent and the server fills it in ([MS-LSAD] 3.1.4.7.2).
type lsarRetrievePrivateDataRequest struct {
	PolicyHandle  mslsad.LSAPR_HANDLE
	KeyName       msdtyp.RPC_UNICODE_STRING
	EncryptedData *mslsad.LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
}

func (*lsarRetrievePrivateDataRequest) Opnum() uint16 { return lsarpc.OpnumLsarRetrievePrivateData }

// lsarRetrievePrivateDataResponse is the reply: the [in,out] encrypted data populated by
// the server followed by the NTSTATUS return value.
type lsarRetrievePrivateDataResponse struct {
	EncryptedData *mslsad.LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
	Status        ndr.DWORD                     `ndr:"retval"`
}

// LsarRetrievePrivateData calls LsarRetrievePrivateData (opnum 43) and returns the
// encrypted data stored under the given key name ([MS-LSAD] 3.1.4.7.2). The result may be
// nil if the server did not return a value.
func LsarRetrievePrivateData(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, keyName string) (*mslsad.LSAPR_CR_CIPHER_VALUE, error) {
	req := &lsarRetrievePrivateDataRequest{
		PolicyHandle: policyHandle,
		KeyName:      msdtyp.NewUnicodeString(keyName),
	}
	var resp lsarRetrievePrivateDataResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("LsarRetrievePrivateData: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.EncryptedData, fmt.Errorf("LsarRetrievePrivateData failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.EncryptedData, nil
}
