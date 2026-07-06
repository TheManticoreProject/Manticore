package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrLogonSendToSamRequest carries the [in] parameters of NetrLogonSendToSam.
type netrLogonSendToSamRequest struct {
	PrimaryName      *ndr.WSTR `ndr:"unique"`
	ComputerName     ndr.WSTR
	Authenticator    msnrpc.NETLOGON_AUTHENTICATOR
	OpaqueBuffer     []uint8 `ndr:"ref,size_is=OpaqueBufferSize"`
	OpaqueBufferSize ndr.DWORD
}

func (*netrLogonSendToSamRequest) Opnum() uint16 { return logon.OpnumNetrLogonSendToSam }

// netrLogonSendToSamResponse carries the [out] parameters and return value of NetrLogonSendToSam.
type netrLogonSendToSamResponse struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	Status              ndr.DWORD `ndr:"retval"`
}

// NetrLogonSendToSam calls NetrLogonSendToSam (opnum 32) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonSendToSam(rpc ndr.Invoker, primaryName *ndr.WSTR, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, opaqueBuffer []uint8, opaqueBufferSize ndr.DWORD) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, err error) {
	req := &netrLogonSendToSamRequest{
		PrimaryName:      primaryName,
		ComputerName:     computerName,
		Authenticator:    authenticator,
		OpaqueBuffer:     opaqueBuffer,
		OpaqueBufferSize: opaqueBufferSize,
	}
	var resp netrLogonSendToSamResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonSendToSam: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonSendToSam failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
