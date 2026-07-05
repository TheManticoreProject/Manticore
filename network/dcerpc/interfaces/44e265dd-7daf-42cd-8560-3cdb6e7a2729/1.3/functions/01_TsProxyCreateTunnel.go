package functions

import (
	"fmt"

	tsgu "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/44e265dd-7daf-42cd-8560-3cdb6e7a2729/1.3"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstsgu "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsgu"
)

// tsProxyCreateTunnelRequest carries the [in] parameters of TsProxyCreateTunnel.
type tsProxyCreateTunnelRequest struct {
	TSGPacket mstsgu.TSG_PACKET
}

func (*tsProxyCreateTunnelRequest) Opnum() uint16 {
	return tsgu.OpnumTsProxyCreateTunnel
}

// tsProxyCreateTunnelResponse carries the [out] parameters and return value of TsProxyCreateTunnel.
type tsProxyCreateTunnelResponse struct {
	TSGPacketResponse *mstsgu.TSG_PACKET `ndr:"unique"`
	TunnelContext     mstsgu.PTUNNEL_CONTEXT_HANDLE_SERIALIZE
	TunnelId          ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// TsProxyCreateTunnel calls TsProxyCreateTunnel (opnum 1) ([MS-TSGU] — verify the parameter
// modeling and status handling).
func TsProxyCreateTunnel(rpc ndr.Invoker, tSGPacket mstsgu.TSG_PACKET) (TSGPacketResponse *mstsgu.TSG_PACKET, TunnelContext mstsgu.PTUNNEL_CONTEXT_HANDLE_SERIALIZE, TunnelId ndr.DWORD, err error) {
	req := &tsProxyCreateTunnelRequest{
		TSGPacket: tSGPacket,
	}
	var resp tsProxyCreateTunnelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("TsProxyCreateTunnel: %w", err)
		return
	}
	TSGPacketResponse = resp.TSGPacketResponse
	TunnelContext = resp.TunnelContext
	TunnelId = resp.TunnelId
	if uint32(resp.Status) != tsgu.StatusSuccess {
		err = fmt.Errorf("TsProxyCreateTunnel failed: %s", tsgu.StatusString(uint32(resp.Status)))
	}
	return
}
