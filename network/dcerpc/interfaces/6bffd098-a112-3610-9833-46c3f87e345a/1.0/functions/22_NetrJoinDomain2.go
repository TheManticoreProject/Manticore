package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrJoinDomain2Request carries the [in] parameters of NetrJoinDomain2.
type netrJoinDomain2Request struct {
	ServerName       *ndr.WSTR `ndr:"unique"`
	DomainNameParam  ndr.WSTR
	MachineAccountOU *ndr.WSTR                              `ndr:"unique"`
	AccountName      *ndr.WSTR                              `ndr:"unique"`
	Password         *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD `ndr:"unique"`
	Options          ndr.DWORD
}

func (*netrJoinDomain2Request) Opnum() uint16 { return wkssvc.OpnumNetrJoinDomain2 }

// netrJoinDomain2Response carries the [out] parameters and return value of NetrJoinDomain2.
type netrJoinDomain2Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrJoinDomain2 calls NetrJoinDomain2 (opnum 22) ([MS-WKST] 3.2.4).
func NetrJoinDomain2(rpc ndr.Invoker, serverName *ndr.WSTR, domainNameParam ndr.WSTR, machineAccountOU *ndr.WSTR, accountName *ndr.WSTR, password *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD, options ndr.DWORD) (err error) {
	req := &netrJoinDomain2Request{
		ServerName:       serverName,
		DomainNameParam:  domainNameParam,
		MachineAccountOU: machineAccountOU,
		AccountName:      accountName,
		Password:         password,
		Options:          options,
	}
	var resp netrJoinDomain2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrJoinDomain2: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrJoinDomain2 failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
