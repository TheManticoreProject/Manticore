package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationSetInformationRequest carries the [in] parameters of RpcWinStationSetInformation.
type rpcWinStationSetInformationRequest struct {
	HServer                     mststs.SERVER_HANDLE
	LogonId                     ndr.DWORD
	WinStationInformationClass  ndr.DWORD
	PWinStationInformation      []int8 `ndr:"ref,size_is=WinStationInformationLength"`
	WinStationInformationLength ndr.DWORD
}

func (*rpcWinStationSetInformationRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationSetInformation
}

// rpcWinStationSetInformationResponse carries the [out] parameters and return value of RpcWinStationSetInformation.
type rpcWinStationSetInformationResponse struct {
	PResult                ndr.DWORD
	PWinStationInformation []int8    `ndr:"ref,size_is=WinStationInformationLength"`
	Status                 ndr.DWORD `ndr:"retval"`
}

// RpcWinStationSetInformation calls RpcWinStationSetInformation (opnum 6) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationSetInformation(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, logonId ndr.DWORD, winStationInformationClass ndr.DWORD, pWinStationInformation []int8, winStationInformationLength ndr.DWORD) (PResult ndr.DWORD, PWinStationInformation []int8, err error) {
	req := &rpcWinStationSetInformationRequest{
		HServer:                     hServer,
		LogonId:                     logonId,
		WinStationInformationClass:  winStationInformationClass,
		PWinStationInformation:      pWinStationInformation,
		WinStationInformationLength: winStationInformationLength,
	}
	var resp rpcWinStationSetInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationSetInformation: %w", err)
		return
	}
	PResult = resp.PResult
	PWinStationInformation = resp.PWinStationInformation
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationSetInformation failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
