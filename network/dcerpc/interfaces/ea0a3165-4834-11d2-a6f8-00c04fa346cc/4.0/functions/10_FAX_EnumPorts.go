package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_EnumPortsRequest carries the [in] parameters of FAX_EnumPorts.
type fAX_EnumPortsRequest struct {
}

func (*fAX_EnumPortsRequest) Opnum() uint16 { return fax.OpnumFAX_EnumPorts }

// fAX_EnumPortsResponse carries the [out] parameters and return value of FAX_EnumPorts.
type fAX_EnumPortsResponse struct {
	PortBuffer    []byte `ndr:"unique,conformant"`
	BufferSize    ndr.DWORD
	PortsReturned ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// FAX_EnumPorts calls FAX_EnumPorts (opnum 10) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_EnumPorts(rpc ndr.Invoker) (PortBuffer []byte, BufferSize ndr.DWORD, PortsReturned ndr.DWORD, err error) {
	req := &fAX_EnumPortsRequest{}
	var resp fAX_EnumPortsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_EnumPorts: %w", err)
		return
	}
	PortBuffer = resp.PortBuffer
	BufferSize = resp.BufferSize
	PortsReturned = resp.PortsReturned
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_EnumPorts failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
