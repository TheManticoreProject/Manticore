package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrServerStatisticsGetRequest is the [in] parameter set of NetrServerStatisticsGet:
// the [unique] server name, the optional [unique] service name, the level (must be 0)
// and the options flags.
type netrServerStatisticsGetRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Service    *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
	Options    ndr.DWORD
}

func (*netrServerStatisticsGetRequest) Opnum() uint16 { return srvsvc.OpnumNetrServerStatisticsGet }

// netrServerStatisticsGetResponse is the reply: the [out] double pointer to a
// STAT_SERVER_0 (a [unique] referent) and the NET_API_STATUS return value.
type netrServerStatisticsGetResponse struct {
	InfoStruct *structures.STAT_SERVER_0 `ndr:"unique"`
	Status     ndr.DWORD                 `ndr:"retval"`
}

// NetrServerStatisticsGet calls NetrServerStatisticsGet (opnum 24), retrieving the
// operating statistics for the server service ([MS-SRVS] 3.1.4.20).
func NetrServerStatisticsGet(rpc ndr.Invoker, serverName, service string, level, options uint32) (*structures.STAT_SERVER_0, error) {
	req := &netrServerStatisticsGetRequest{
		ServerName: optWStr(serverName),
		Service:    optWStr(service),
		Level:      ndr.DWORD(level),
		Options:    ndr.DWORD(options),
	}
	var resp netrServerStatisticsGetResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("NetrServerStatisticsGet: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success {
		return resp.InfoStruct, fmt.Errorf("NetrServerStatisticsGet failed: %s", srvsvc.StatusString(status))
	}
	return resp.InfoStruct, nil
}
