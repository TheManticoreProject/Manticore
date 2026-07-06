package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrUseAddRequest carries the [in] parameters of NetrUseAdd.
type netrUseAddRequest struct {
	ServerName     *ndr.WSTR `ndr:"unique"`
	Level          ndr.DWORD
	InfoStruct     mswkst.USE_INFO
	ErrorParameter *ndr.DWORD `ndr:"unique"`
}

func (*netrUseAddRequest) Opnum() uint16 { return wkssvc.OpnumNetrUseAdd }

// netrUseAddResponse carries the [out] parameters and return value of NetrUseAdd.
type netrUseAddResponse struct {
	ErrorParameter *ndr.DWORD `ndr:"unique"`
	Status         ndr.DWORD  `ndr:"retval"`
}

// NetrUseAdd calls NetrUseAdd (opnum 8) ([MS-WKST] 3.2.4).
func NetrUseAdd(rpc ndr.Invoker, serverName *ndr.WSTR, level ndr.DWORD, infoStruct mswkst.USE_INFO, errorParameter *ndr.DWORD) (ErrorParameter *ndr.DWORD, err error) {
	req := &netrUseAddRequest{
		ServerName:     serverName,
		Level:          level,
		InfoStruct:     infoStruct,
		ErrorParameter: errorParameter,
	}
	var resp netrUseAddResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrUseAdd: %w", err)
		return
	}
	ErrorParameter = resp.ErrorParameter
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrUseAdd failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
