package functions

// IDL source: [MS-TSGU] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsgu/ea0ac9e8-2d53-477e-ba57-b1ad01e38039
// A fetched copy is kept at ms-tsgu.idl in the interface directory.

import (
	"fmt"

	tsgu "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/44e265dd-7daf-42cd-8560-3cdb6e7a2729/1.3"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstsgu "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsgu"
)

// tsProxyCreateChannelRequest carries the [in] parameters of TsProxyCreateChannel.
type tsProxyCreateChannelRequest struct {
	TunnelContext  mstsgu.PTUNNEL_CONTEXT_HANDLE_NOSERIALIZE
	TsEndPointInfo mstsgu.TSENDPOINTINFO
}

func (*tsProxyCreateChannelRequest) Opnum() uint16 {
	return tsgu.OpnumTsProxyCreateChannel
}

// tsProxyCreateChannelResponse carries the [out] parameters and return value of TsProxyCreateChannel.
type tsProxyCreateChannelResponse struct {
	ChannelContext mstsgu.PCHANNEL_CONTEXT_HANDLE_SERIALIZE
	ChannelId      ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// TsProxyCreateChannel calls TsProxyCreateChannel (opnum 4) ([MS-TSGU] — verify the parameter
// modeling and status handling).
func TsProxyCreateChannel(rpc ndr.Invoker, tunnelContext mstsgu.PTUNNEL_CONTEXT_HANDLE_NOSERIALIZE, tsEndPointInfo mstsgu.TSENDPOINTINFO) (ChannelContext mstsgu.PCHANNEL_CONTEXT_HANDLE_SERIALIZE, ChannelId ndr.DWORD, err error) {
	req := &tsProxyCreateChannelRequest{
		TunnelContext:  tunnelContext,
		TsEndPointInfo: tsEndPointInfo,
	}
	var resp tsProxyCreateChannelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("TsProxyCreateChannel: %w", err)
		return
	}
	ChannelContext = resp.ChannelContext
	ChannelId = resp.ChannelId
	if uint32(resp.Status) != tsgu.StatusSuccess {
		err = fmt.Errorf("TsProxyCreateChannel failed: %s", tsgu.StatusString(uint32(resp.Status)))
	}
	return
}
