package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrRenameMachineInDomain2Request carries the [in] parameters of NetrRenameMachineInDomain2.
type netrRenameMachineInDomain2Request struct {
	ServerName  *ndr.WSTR                              `ndr:"unique"`
	MachineName *ndr.WSTR                              `ndr:"unique"`
	AccountName *ndr.WSTR                              `ndr:"unique"`
	Password    *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD `ndr:"unique"`
	Options     ndr.DWORD
}

func (*netrRenameMachineInDomain2Request) Opnum() uint16 {
	return wkssvc.OpnumNetrRenameMachineInDomain2
}

// netrRenameMachineInDomain2Response carries the [out] parameters and return value of NetrRenameMachineInDomain2.
type netrRenameMachineInDomain2Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrRenameMachineInDomain2 calls NetrRenameMachineInDomain2 (opnum 24) ([MS-WKST] 3.2.4).
func NetrRenameMachineInDomain2(rpc ndr.Invoker, serverName *ndr.WSTR, machineName *ndr.WSTR, accountName *ndr.WSTR, password *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD, options ndr.DWORD) (err error) {
	req := &netrRenameMachineInDomain2Request{
		ServerName:  serverName,
		MachineName: machineName,
		AccountName: accountName,
		Password:    password,
		Options:     options,
	}
	var resp netrRenameMachineInDomain2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrRenameMachineInDomain2: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrRenameMachineInDomain2 failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
