package functions

// IDL source: [MS-TSCH] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsch/6fc1f51a-26ec-43fa-a8bd-1c364657f110
// A fetched copy is kept at ms-tsch.idl in the interface directory.

import (
	"fmt"

	atsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1ff70682-0a51-30e8-076d-740be8cee98b/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrJobDelRequest carries the [in] parameters of NetrJobDel.
type netrJobDelRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	MinJobId   ndr.DWORD
	MaxJobId   ndr.DWORD
}

func (*netrJobDelRequest) Opnum() uint16 { return atsvc.OpnumNetrJobDel }

// netrJobDelResponse carries the [out] parameters and return value of NetrJobDel.
type netrJobDelResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrJobDel calls NetrJobDel (opnum 1) ([MS-TSCH] section 3.2.5.2.2).
func NetrJobDel(rpc ndr.Invoker, serverName *ndr.WSTR, minJobId ndr.DWORD, maxJobId ndr.DWORD) (err error) {
	req := &netrJobDelRequest{
		ServerName: serverName,
		MinJobId:   minJobId,
		MaxJobId:   maxJobId,
	}
	var resp netrJobDelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrJobDel: %w", err)
		return
	}
	if uint32(resp.Status) != atsvc.StatusSuccess {
		err = fmt.Errorf("NetrJobDel failed: %s", atsvc.StatusString(uint32(resp.Status)))
	}
	return
}
