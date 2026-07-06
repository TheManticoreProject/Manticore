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

// netrLogonControlRequest carries the [in] parameters of NetrLogonControl.
type netrLogonControlRequest struct {
	ServerName   *ndr.WSTR `ndr:"unique"`
	FunctionCode ndr.DWORD
	QueryLevel   ndr.DWORD
}

func (*netrLogonControlRequest) Opnum() uint16 { return logon.OpnumNetrLogonControl }

// netrLogonControlResponse carries the [out] parameters and return value of NetrLogonControl.
type netrLogonControlResponse struct {
	Buffer msnrpc.NETLOGON_CONTROL_QUERY_INFORMATION
	Status ndr.DWORD `ndr:"retval"`
}

// NetrLogonControl calls NetrLogonControl (opnum 12) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonControl(rpc ndr.Invoker, serverName *ndr.WSTR, functionCode ndr.DWORD, queryLevel ndr.DWORD) (Buffer msnrpc.NETLOGON_CONTROL_QUERY_INFORMATION, err error) {
	req := &netrLogonControlRequest{
		ServerName:   serverName,
		FunctionCode: functionCode,
		QueryLevel:   queryLevel,
	}
	var resp netrLogonControlResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonControl: %w", err)
		return
	}
	Buffer = resp.Buffer
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonControl failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
