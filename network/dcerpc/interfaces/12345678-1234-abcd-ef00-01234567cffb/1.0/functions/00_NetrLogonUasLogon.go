package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrLogonUasLogonRequest carries the [in] parameters of NetrLogonUasLogon.
type netrLogonUasLogonRequest struct {
	ServerName  *ndr.WSTR `ndr:"unique"`
	UserName    ndr.WSTR
	Workstation ndr.WSTR
}

func (*netrLogonUasLogonRequest) Opnum() uint16 { return logon.OpnumNetrLogonUasLogon }

// netrLogonUasLogonResponse carries the [out] parameters and return value of NetrLogonUasLogon.
type netrLogonUasLogonResponse struct {
	ValidationInformation *msnrpc.NETLOGON_VALIDATION_UAS_INFO `ndr:"unique"`
	Status                ndr.DWORD                            `ndr:"retval"`
}

// NetrLogonUasLogon calls NetrLogonUasLogon (opnum 0) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonUasLogon(rpc ndr.Invoker, serverName *ndr.WSTR, userName ndr.WSTR, workstation ndr.WSTR) (ValidationInformation *msnrpc.NETLOGON_VALIDATION_UAS_INFO, err error) {
	req := &netrLogonUasLogonRequest{
		ServerName:  serverName,
		UserName:    userName,
		Workstation: workstation,
	}
	var resp netrLogonUasLogonResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonUasLogon: %w", err)
		return
	}
	ValidationInformation = resp.ValidationInformation
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonUasLogon failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
