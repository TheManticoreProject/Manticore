package functions

// IDL source: [MS-EVEN6] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-even6/2d808edd-719a-4c69-b34a-df766adb5f0c
// A fetched copy is kept at ms-even6.idl in the interface directory.

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcQuerySeekRequest carries the [in] parameters of EvtRpcQuerySeek.
type evtRpcQuerySeekRequest struct {
	LogQuery    mseven6.PCONTEXT_HANDLE_LOG_QUERY
	Pos         int64
	BookmarkXml *ndr.WSTR `ndr:"unique"`
	TimeOut     ndr.DWORD
	Flags       ndr.DWORD
}

func (*evtRpcQuerySeekRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcQuerySeek }

// evtRpcQuerySeekResponse carries the [out] parameters and return value of EvtRpcQuerySeek.
type evtRpcQuerySeekResponse struct {
	Error  mseven6.RpcInfo
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcQuerySeek calls EvtRpcQuerySeek (opnum 12) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcQuerySeek(rpc ndr.Invoker, logQuery mseven6.PCONTEXT_HANDLE_LOG_QUERY, pos int64, bookmarkXml *ndr.WSTR, timeOut ndr.DWORD, flags ndr.DWORD) (Error mseven6.RpcInfo, err error) {
	req := &evtRpcQuerySeekRequest{
		LogQuery:    logQuery,
		Pos:         pos,
		BookmarkXml: bookmarkXml,
		TimeOut:     timeOut,
		Flags:       flags,
	}
	var resp evtRpcQuerySeekResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcQuerySeek: %w", err)
		return
	}
	Error = resp.Error
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcQuerySeek failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
