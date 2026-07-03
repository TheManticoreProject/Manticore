package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrLogonControl2ExRequest carries the [in] parameters of NetrLogonControl2Ex.
type netrLogonControl2ExRequest struct {
	ServerName   *ndr.WSTR `ndr:"unique"`
	FunctionCode ndr.DWORD
	QueryLevel   ndr.DWORD
	Data         msnrpc.NETLOGON_CONTROL_DATA_INFORMATION
}

func (*netrLogonControl2ExRequest) Opnum() uint16 { return logon.OpnumNetrLogonControl2Ex }

// netrLogonControl2ExResponse carries the [out] parameters and return value of NetrLogonControl2Ex.
type netrLogonControl2ExResponse struct {
	Buffer msnrpc.NETLOGON_CONTROL_QUERY_INFORMATION
	Status ndr.DWORD `ndr:"retval"`
}

// NetrLogonControl2Ex calls NetrLogonControl2Ex (opnum 18) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonControl2Ex(rpc ndr.Invoker, serverName *ndr.WSTR, functionCode ndr.DWORD, queryLevel ndr.DWORD, data msnrpc.NETLOGON_CONTROL_DATA_INFORMATION) (Buffer msnrpc.NETLOGON_CONTROL_QUERY_INFORMATION, err error) {
	req := &netrLogonControl2ExRequest{
		ServerName:   serverName,
		FunctionCode: functionCode,
		QueryLevel:   queryLevel,
		Data:         data,
	}
	var resp netrLogonControl2ExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonControl2Ex: %w", err)
		return
	}
	Buffer = resp.Buffer
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonControl2Ex failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
