package functions

// IDL source: [MS-WKST] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-wkst/9fdbc753-0397-4236-bbfc-a380f9d23789
// A fetched copy is kept at ms-wkst.idl in the interface directory.

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrWorkstationStatisticsGetRequest carries the [in] parameters of NetrWorkstationStatisticsGet.
type netrWorkstationStatisticsGetRequest struct {
	ServerName  *ndr.WSTR `ndr:"unique"`
	ServiceName *ndr.WSTR `ndr:"unique"`
	Level       ndr.DWORD
	Options     ndr.DWORD
}

func (*netrWorkstationStatisticsGetRequest) Opnum() uint16 {
	return wkssvc.OpnumNetrWorkstationStatisticsGet
}

// netrWorkstationStatisticsGetResponse carries the [out] parameters and return value of NetrWorkstationStatisticsGet.
type netrWorkstationStatisticsGetResponse struct {
	Buffer *mswkst.STAT_WORKSTATION_0 `ndr:"unique"`
	Status ndr.DWORD                  `ndr:"retval"`
}

// NetrWorkstationStatisticsGet calls NetrWorkstationStatisticsGet (opnum 13) ([MS-WKST] 3.2.4).
func NetrWorkstationStatisticsGet(rpc ndr.Invoker, serverName *ndr.WSTR, serviceName *ndr.WSTR, level ndr.DWORD, options ndr.DWORD) (Buffer *mswkst.STAT_WORKSTATION_0, err error) {
	req := &netrWorkstationStatisticsGetRequest{
		ServerName:  serverName,
		ServiceName: serviceName,
		Level:       level,
		Options:     options,
	}
	var resp netrWorkstationStatisticsGetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrWorkstationStatisticsGet: %w", err)
		return
	}
	Buffer = resp.Buffer
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrWorkstationStatisticsGet failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
