package functions

// IDL source: [MS-TSCH] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsch/6fc1f51a-26ec-43fa-a8bd-1c364657f110
// A fetched copy is kept at ms-tsch.idl in the interface directory.

import (
	"fmt"

	atsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1ff70682-0a51-30e8-076d-740be8cee98b/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstsch "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsch"
)

// netrJobAddRequest carries the [in] parameters of NetrJobAdd.
type netrJobAddRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	PAtInfo    mstsch.AT_INFO
}

func (*netrJobAddRequest) Opnum() uint16 { return atsvc.OpnumNetrJobAdd }

// netrJobAddResponse carries the [out] parameters and return value of NetrJobAdd.
type netrJobAddResponse struct {
	PJobId ndr.DWORD
	Status ndr.DWORD `ndr:"retval"`
}

// NetrJobAdd calls NetrJobAdd (opnum 0) ([MS-TSCH] section 3.2.5.2.1).
func NetrJobAdd(rpc ndr.Invoker, serverName *ndr.WSTR, pAtInfo mstsch.AT_INFO) (PJobId ndr.DWORD, err error) {
	req := &netrJobAddRequest{
		ServerName: serverName,
		PAtInfo:    pAtInfo,
	}
	var resp netrJobAddResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrJobAdd: %w", err)
		return
	}
	PJobId = resp.PJobId
	if uint32(resp.Status) != atsvc.StatusSuccess {
		err = fmt.Errorf("NetrJobAdd failed: %s", atsvc.StatusString(uint32(resp.Status)))
	}
	return
}
