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

// netrLogonUasLogoffRequest carries the [in] parameters of NetrLogonUasLogoff.
type netrLogonUasLogoffRequest struct {
	ServerName  *ndr.WSTR `ndr:"unique"`
	UserName    ndr.WSTR
	Workstation ndr.WSTR
}

func (*netrLogonUasLogoffRequest) Opnum() uint16 { return logon.OpnumNetrLogonUasLogoff }

// netrLogonUasLogoffResponse carries the [out] parameters and return value of NetrLogonUasLogoff.
type netrLogonUasLogoffResponse struct {
	LogoffInformation msnrpc.NETLOGON_LOGOFF_UAS_INFO
	Status            ndr.DWORD `ndr:"retval"`
}

// NetrLogonUasLogoff calls NetrLogonUasLogoff (opnum 1) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonUasLogoff(rpc ndr.Invoker, serverName *ndr.WSTR, userName ndr.WSTR, workstation ndr.WSTR) (LogoffInformation msnrpc.NETLOGON_LOGOFF_UAS_INFO, err error) {
	req := &netrLogonUasLogoffRequest{
		ServerName:  serverName,
		UserName:    userName,
		Workstation: workstation,
	}
	var resp netrLogonUasLogoffResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonUasLogoff: %w", err)
		return
	}
	LogoffInformation = resp.LogoffInformation
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonUasLogoff failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
