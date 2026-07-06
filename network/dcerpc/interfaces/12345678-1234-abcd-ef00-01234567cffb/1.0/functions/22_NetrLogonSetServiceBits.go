package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrLogonSetServiceBitsRequest carries the [in] parameters of NetrLogonSetServiceBits.
type netrLogonSetServiceBitsRequest struct {
	ServerName            *ndr.WSTR `ndr:"unique"`
	ServiceBitsOfInterest ndr.DWORD
	ServiceBits           ndr.DWORD
}

func (*netrLogonSetServiceBitsRequest) Opnum() uint16 { return logon.OpnumNetrLogonSetServiceBits }

// netrLogonSetServiceBitsResponse carries the [out] parameters and return value of NetrLogonSetServiceBits.
type netrLogonSetServiceBitsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrLogonSetServiceBits calls NetrLogonSetServiceBits (opnum 22) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonSetServiceBits(rpc ndr.Invoker, serverName *ndr.WSTR, serviceBitsOfInterest ndr.DWORD, serviceBits ndr.DWORD) (err error) {
	req := &netrLogonSetServiceBitsRequest{
		ServerName:            serverName,
		ServiceBitsOfInterest: serviceBitsOfInterest,
		ServiceBits:           serviceBits,
	}
	var resp netrLogonSetServiceBitsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonSetServiceBits: %w", err)
		return
	}
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonSetServiceBits failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
