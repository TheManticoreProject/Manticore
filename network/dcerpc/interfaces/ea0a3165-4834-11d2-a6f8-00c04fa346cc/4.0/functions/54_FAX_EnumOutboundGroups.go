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

// fAX_EnumOutboundGroupsRequest carries the [in] parameters of FAX_EnumOutboundGroups.
type fAX_EnumOutboundGroupsRequest struct {
}

func (*fAX_EnumOutboundGroupsRequest) Opnum() uint16 { return fax.OpnumFAX_EnumOutboundGroups }

// fAX_EnumOutboundGroupsResponse carries the [out] parameters and return value of FAX_EnumOutboundGroups.
type fAX_EnumOutboundGroupsResponse struct {
	PpData        []byte `ndr:"unique,conformant"`
	LpdwDataSize  ndr.DWORD
	LpdwNumGroups ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// FAX_EnumOutboundGroups calls FAX_EnumOutboundGroups (opnum 54) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_EnumOutboundGroups(rpc ndr.Invoker) (PpData []byte, LpdwDataSize ndr.DWORD, LpdwNumGroups ndr.DWORD, err error) {
	req := &fAX_EnumOutboundGroupsRequest{}
	var resp fAX_EnumOutboundGroupsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_EnumOutboundGroups: %w", err)
		return
	}
	PpData = resp.PpData
	LpdwDataSize = resp.LpdwDataSize
	LpdwNumGroups = resp.LpdwNumGroups
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_EnumOutboundGroups failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
