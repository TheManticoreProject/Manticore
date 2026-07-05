package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrServerTransportDelExRequest is the [in] parameter set of NetrServerTransportDelEx:
// the [unique] server name, the level, and the [in, switch_is(Level)] TRANSPORT_INFO
// union (its Tag is set to Level before marshalling).
type netrServerTransportDelExRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
	Buffer     mssrvs.TRANSPORT_INFO
}

func (*netrServerTransportDelExRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrServerTransportDelEx
}

// NetrServerTransportDelEx calls NetrServerTransportDelEx (opnum 53), unbinding the
// server from a transport protocol with extended (leveled) information ([MS-SRVS]
// 3.1.4.26).
func NetrServerTransportDelEx(rpc ndr.Invoker, serverName string, level uint32, buffer mssrvs.TRANSPORT_INFO) error {
	buffer.Tag = ndr.DWORD(level)
	req := &netrServerTransportDelExRequest{
		ServerName: optWStr(serverName),
		Level:      ndr.DWORD(level),
		Buffer:     buffer,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrServerTransportDelEx: %w", err)
	}
	if status := uint32(resp.Status); status != srvsvc.NERR_Success {
		return fmt.Errorf("NetrServerTransportDelEx failed: %s", srvsvc.StatusString(status))
	}
	return nil
}
