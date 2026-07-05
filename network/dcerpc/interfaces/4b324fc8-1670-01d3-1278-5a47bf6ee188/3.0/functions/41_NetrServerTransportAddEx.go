package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrServerTransportAddExRequest is the [in] parameter set of NetrServerTransportAddEx:
// the [unique] server name, the level, and the [in, switch_is(Level)] TRANSPORT_INFO
// union (its Tag is set to Level before marshalling).
type netrServerTransportAddExRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
	Buffer     mssrvs.TRANSPORT_INFO
}

func (*netrServerTransportAddExRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrServerTransportAddEx
}

// NetrServerTransportAddEx calls NetrServerTransportAddEx (opnum 41), binding the
// server to a transport protocol with extended (leveled) information ([MS-SRVS]
// 3.1.4.25).
func NetrServerTransportAddEx(rpc ndr.Invoker, serverName string, level uint32, buffer mssrvs.TRANSPORT_INFO) error {
	buffer.Tag = ndr.DWORD(level)
	req := &netrServerTransportAddExRequest{
		ServerName: optWStr(serverName),
		Level:      ndr.DWORD(level),
		Buffer:     buffer,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrServerTransportAddEx: %w", err)
	}
	if status := uint32(resp.Status); status != srvsvc.NERR_Success {
		return fmt.Errorf("NetrServerTransportAddEx failed: %s", srvsvc.StatusString(status))
	}
	return nil
}
