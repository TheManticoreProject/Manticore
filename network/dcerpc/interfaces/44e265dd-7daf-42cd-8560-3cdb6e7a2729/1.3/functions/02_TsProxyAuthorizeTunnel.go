package functions

import (
	"fmt"

	tsgu "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/44e265dd-7daf-42cd-8560-3cdb6e7a2729/1.3"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstsgu "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsgu"
)

// tsProxyAuthorizeTunnelRequest carries the [in] parameters of TsProxyAuthorizeTunnel.
type tsProxyAuthorizeTunnelRequest struct {
	TunnelContext mstsgu.PTUNNEL_CONTEXT_HANDLE_NOSERIALIZE
	TSGPacket     mstsgu.TSG_PACKET
}

func (*tsProxyAuthorizeTunnelRequest) Opnum() uint16 {
	return tsgu.OpnumTsProxyAuthorizeTunnel
}

// tsProxyAuthorizeTunnelResponse carries the [out] parameters and return value of TsProxyAuthorizeTunnel.
type tsProxyAuthorizeTunnelResponse struct {
	TSGPacketResponse *mstsgu.TSG_PACKET `ndr:"unique"`
	Status            ndr.DWORD          `ndr:"retval"`
}

// TsProxyAuthorizeTunnel calls TsProxyAuthorizeTunnel (opnum 2) ([MS-TSGU] — verify the parameter
// modeling and status handling).
func TsProxyAuthorizeTunnel(rpc ndr.Invoker, tunnelContext mstsgu.PTUNNEL_CONTEXT_HANDLE_NOSERIALIZE, tSGPacket mstsgu.TSG_PACKET) (TSGPacketResponse *mstsgu.TSG_PACKET, err error) {
	req := &tsProxyAuthorizeTunnelRequest{
		TunnelContext: tunnelContext,
		TSGPacket:     tSGPacket,
	}
	var resp tsProxyAuthorizeTunnelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("TsProxyAuthorizeTunnel: %w", err)
		return
	}
	TSGPacketResponse = resp.TSGPacketResponse
	if uint32(resp.Status) != tsgu.StatusSuccess {
		err = fmt.Errorf("TsProxyAuthorizeTunnel failed: %s", tsgu.StatusString(uint32(resp.Status)))
	}
	return
}
