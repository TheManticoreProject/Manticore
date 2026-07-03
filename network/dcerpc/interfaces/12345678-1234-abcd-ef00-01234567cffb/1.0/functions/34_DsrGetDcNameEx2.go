package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// dsrGetDcNameEx2Request carries the [in] parameters of DsrGetDcNameEx2.
type dsrGetDcNameEx2Request struct {
	ComputerName                *ndr.WSTR `ndr:"unique"`
	AccountName                 *ndr.WSTR `ndr:"unique"`
	AllowableAccountControlBits ndr.DWORD
	DomainName                  *ndr.WSTR  `ndr:"unique"`
	DomainGuid                  *guid.GUID `ndr:"unique"`
	SiteName                    *ndr.WSTR  `ndr:"unique"`
	Flags                       ndr.DWORD
}

func (*dsrGetDcNameEx2Request) Opnum() uint16 { return logon.OpnumDsrGetDcNameEx2 }

// dsrGetDcNameEx2Response carries the [out] parameters and return value of DsrGetDcNameEx2.
type dsrGetDcNameEx2Response struct {
	DomainControllerInfo *msnrpc.DOMAIN_CONTROLLER_INFOW `ndr:"unique"`
	Status               ndr.DWORD                       `ndr:"retval"`
}

// DsrGetDcNameEx2 calls DsrGetDcNameEx2 (opnum 34) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func DsrGetDcNameEx2(rpc ndr.Invoker, computerName *ndr.WSTR, accountName *ndr.WSTR, allowableAccountControlBits ndr.DWORD, domainName *ndr.WSTR, domainGuid *guid.GUID, siteName *ndr.WSTR, flags ndr.DWORD) (DomainControllerInfo *msnrpc.DOMAIN_CONTROLLER_INFOW, err error) {
	req := &dsrGetDcNameEx2Request{
		ComputerName:                computerName,
		AccountName:                 accountName,
		AllowableAccountControlBits: allowableAccountControlBits,
		DomainName:                  domainName,
		DomainGuid:                  domainGuid,
		SiteName:                    siteName,
		Flags:                       flags,
	}
	var resp dsrGetDcNameEx2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DsrGetDcNameEx2: %w", err)
		return
	}
	DomainControllerInfo = resp.DomainControllerInfo
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("DsrGetDcNameEx2 failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
