package functions

import (
	"fmt"

	msgsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/17fdd703-1827-4e34-79d4-24a55c53bb37/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmsrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-msrp"
)

// netrMessageNameGetInfoRequest carries the [in] parameters of NetrMessageNameGetInfo.
type netrMessageNameGetInfoRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	MsgName    ndr.WSTR
	Level      ndr.DWORD
}

func (*netrMessageNameGetInfoRequest) Opnum() uint16 { return msgsvc.OpnumNetrMessageNameGetInfo }

// netrMessageNameGetInfoResponse carries the [out] parameters and return value of NetrMessageNameGetInfo.
type netrMessageNameGetInfoResponse struct {
	InfoStruct msmsrp.MSG_INFO
	Status     ndr.DWORD `ndr:"retval"`
}

// NetrMessageNameGetInfo calls NetrMessageNameGetInfo (opnum 2) ([MS-MSRP] — verify the parameter
// modeling and status handling).
func NetrMessageNameGetInfo(rpc ndr.Invoker, serverName *ndr.WSTR, msgName ndr.WSTR, level ndr.DWORD) (InfoStruct msmsrp.MSG_INFO, err error) {
	req := &netrMessageNameGetInfoRequest{
		ServerName: serverName,
		MsgName:    msgName,
		Level:      level,
	}
	var resp netrMessageNameGetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrMessageNameGetInfo: %w", err)
		return
	}
	InfoStruct = resp.InfoStruct
	if uint32(resp.Status) != msgsvc.StatusSuccess {
		err = fmt.Errorf("NetrMessageNameGetInfo failed: %s", msgsvc.StatusString(uint32(resp.Status)))
	}
	return
}
