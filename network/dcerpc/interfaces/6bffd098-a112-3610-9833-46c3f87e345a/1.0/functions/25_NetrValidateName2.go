package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrValidateName2Request carries the [in] parameters of NetrValidateName2.
type netrValidateName2Request struct {
	ServerName     *ndr.WSTR `ndr:"unique"`
	NameToValidate ndr.WSTR
	AccountName    *ndr.WSTR                              `ndr:"unique"`
	Password       *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD `ndr:"unique"`
	NameType       mswkst.NETSETUP_NAME_TYPE
}

func (*netrValidateName2Request) Opnum() uint16 { return wkssvc.OpnumNetrValidateName2 }

// netrValidateName2Response carries the [out] parameters and return value of NetrValidateName2.
type netrValidateName2Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrValidateName2 calls NetrValidateName2 (opnum 25) ([MS-WKST] 3.2.4).
func NetrValidateName2(rpc ndr.Invoker, serverName *ndr.WSTR, nameToValidate ndr.WSTR, accountName *ndr.WSTR, password *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD, nameType mswkst.NETSETUP_NAME_TYPE) (err error) {
	req := &netrValidateName2Request{
		ServerName:     serverName,
		NameToValidate: nameToValidate,
		AccountName:    accountName,
		Password:       password,
		NameType:       nameType,
	}
	var resp netrValidateName2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrValidateName2: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrValidateName2 failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
