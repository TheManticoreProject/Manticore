package functions

// IDL source: [MS-DFSNM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dfsnm/b471e023-618d-4c48-877f-f30c3005320c
// A fetched copy is kept at ms-dfsnm.idl in the interface directory.

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrDfsManagerInitializeRequest carries the [in] parameters of NetrDfsManagerInitialize.
type netrDfsManagerInitializeRequest struct {
	ServerName ndr.WSTR
	Flags      ndr.DWORD
}

func (*netrDfsManagerInitializeRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsManagerInitialize }

// netrDfsManagerInitializeResponse carries the [out] parameters and return value of NetrDfsManagerInitialize.
type netrDfsManagerInitializeResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrDfsManagerInitialize calls NetrDfsManagerInitialize (opnum 14) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsManagerInitialize(rpc ndr.Invoker, serverName ndr.WSTR, flags ndr.DWORD) (err error) {
	req := &netrDfsManagerInitializeRequest{
		ServerName: serverName,
		Flags:      flags,
	}
	var resp netrDfsManagerInitializeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsManagerInitialize: %w", err)
		return
	}
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsManagerInitialize failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
