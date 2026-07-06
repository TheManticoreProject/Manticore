package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrEnumerateComputerNamesRequest carries the [in] parameters of NetrEnumerateComputerNames.
type netrEnumerateComputerNamesRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	NameType   mswkst.NET_COMPUTER_NAME_TYPE
	Reserved   ndr.DWORD
}

func (*netrEnumerateComputerNamesRequest) Opnum() uint16 {
	return wkssvc.OpnumNetrEnumerateComputerNames
}

// netrEnumerateComputerNamesResponse carries the [out] parameters and return value of NetrEnumerateComputerNames.
type netrEnumerateComputerNamesResponse struct {
	ComputerNames *mswkst.NET_COMPUTER_NAME_ARRAY `ndr:"unique"`
	Status        ndr.DWORD                       `ndr:"retval"`
}

// NetrEnumerateComputerNames calls NetrEnumerateComputerNames (opnum 30) ([MS-WKST] 3.2.4).
func NetrEnumerateComputerNames(rpc ndr.Invoker, serverName *ndr.WSTR, nameType mswkst.NET_COMPUTER_NAME_TYPE, reserved ndr.DWORD) (ComputerNames *mswkst.NET_COMPUTER_NAME_ARRAY, err error) {
	req := &netrEnumerateComputerNamesRequest{
		ServerName: serverName,
		NameType:   nameType,
		Reserved:   reserved,
	}
	var resp netrEnumerateComputerNamesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrEnumerateComputerNames: %w", err)
		return
	}
	ComputerNames = resp.ComputerNames
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrEnumerateComputerNames failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
