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

// netrJobGetInfoRequest carries the [in] parameters of NetrJobGetInfo.
type netrJobGetInfoRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	JobId      ndr.DWORD
}

func (*netrJobGetInfoRequest) Opnum() uint16 { return atsvc.OpnumNetrJobGetInfo }

// netrJobGetInfoResponse carries the [out] parameters and return value of NetrJobGetInfo.
type netrJobGetInfoResponse struct {
	PpAtInfo *mstsch.AT_INFO `ndr:"unique"`
	Status   ndr.DWORD       `ndr:"retval"`
}

// NetrJobGetInfo calls NetrJobGetInfo (opnum 3) ([MS-TSCH] section 3.2.5.2.4).
func NetrJobGetInfo(rpc ndr.Invoker, serverName *ndr.WSTR, jobId ndr.DWORD) (PpAtInfo *mstsch.AT_INFO, err error) {
	req := &netrJobGetInfoRequest{
		ServerName: serverName,
		JobId:      jobId,
	}
	var resp netrJobGetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrJobGetInfo: %w", err)
		return
	}
	PpAtInfo = resp.PpAtInfo
	if uint32(resp.Status) != atsvc.StatusSuccess {
		err = fmt.Errorf("NetrJobGetInfo failed: %s", atsvc.StatusString(uint32(resp.Status)))
	}
	return
}
